package handlers

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newStrategyHandler(t *testing.T) (*StrategyHandler, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	return NewStrategyHandler(sdb, nil, zap.NewNop()), mock
}

var strategyColumns = []string{"id", "user_id", "name", "config", "created_at", "updated_at"}

func TestGetStrategies(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`FROM strategies`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(strategyColumns).
			AddRow(uuid.New(), userID, "Dip", []byte(`{"entry":{}}`), time.Now(), time.Now()))

	resp, err := h.GetStrategies(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	require.Len(t, resp.Strategies, 1)
	assert.Equal(t, "Dip", resp.Strategies[0].Name)
}

func TestGetStrategiesInvalidUser(t *testing.T) {
	h, _ := newStrategyHandler(t)
	_, err := h.GetStrategies(context.Background(), &userpb.UserSpecifier{UserId: "bad"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestCreateStrategy(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()

	mock.ExpectQuery(`INSERT INTO strategies`).
		WithArgs(userID, "Dip", []byte(`{"entry":{}}`)).
		WillReturnRows(sqlmock.NewRows(strategyColumns).
			AddRow(stratID, userID, "Dip", []byte(`{"entry":{}}`), time.Now(), time.Now()))

	resp, err := h.CreateStrategy(context.Background(), &userpb.CreateStrategyRequest{
		UserId: userID.String(), Name: "Dip", ConfigJson: `{"entry":{}}`,
	})
	require.NoError(t, err)
	assert.Equal(t, stratID.String(), resp.Id)
}

func TestCreateStrategyRejectsBadJSON(t *testing.T) {
	h, _ := newStrategyHandler(t)
	_, err := h.CreateStrategy(context.Background(), &userpb.CreateStrategyRequest{
		UserId: uuid.New().String(), Name: "X", ConfigJson: "{not json",
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestUpdateStrategy(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()

	mock.ExpectQuery(`UPDATE strategies`).
		WithArgs(stratID, userID, "New", []byte(`{"entry":{}}`)).
		WillReturnRows(sqlmock.NewRows(strategyColumns).
			AddRow(stratID, userID, "New", []byte(`{"entry":{}}`), time.Now(), time.Now()))

	resp, err := h.UpdateStrategy(context.Background(), &userpb.UpdateStrategyRequest{
		Id: stratID.String(), UserId: userID.String(), Name: "New", ConfigJson: `{"entry":{}}`,
	})
	require.NoError(t, err)
	assert.Equal(t, "New", resp.Name)
}

func TestUpdateStrategyNotFound(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()

	mock.ExpectQuery(`UPDATE strategies`).WillReturnError(sql.ErrNoRows)

	_, err := h.UpdateStrategy(context.Background(), &userpb.UpdateStrategyRequest{
		Id: stratID.String(), UserId: userID.String(), Name: "New", ConfigJson: `{}`,
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteStrategy(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()

	mock.ExpectExec(`DELETE FROM strategies`).
		WithArgs(stratID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.DeleteStrategy(context.Background(), &userpb.StrategySpecifier{
		Id: stratID.String(), UserId: userID.String(),
	})
	require.NoError(t, err)
}

func TestDeleteStrategyInvalidID(t *testing.T) {
	h, _ := newStrategyHandler(t)
	_, err := h.DeleteStrategy(context.Background(), &userpb.StrategySpecifier{
		Id: "bad", UserId: uuid.New().String(),
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

var deploymentColumns = []string{
	"id", "user_id", "strategy_id", "symbol", "position_size_cents", "status",
	"in_position", "entry_price_cents", "qty", "created_at", "updated_at",
}

func TestDeployStrategy(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()
	depID := uuid.New()
	cfg := []byte(`{"entry":{"rules":[{"lhs":{"kind":"price"},"op":">","rhs":{"kind":"value","value":1}}]}}`)

	mock.ExpectQuery(`FROM strategies`).
		WithArgs(stratID, userID).
		WillReturnRows(sqlmock.NewRows(strategyColumns).
			AddRow(stratID, userID, "Dip", cfg, time.Now(), time.Now()))
	mock.ExpectQuery(`FROM strategy_deployments`).
		WithArgs(userID, stratID, "AAPL").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectQuery(`INSERT INTO strategy_deployments`).
		WithArgs(userID, stratID, "AAPL", int64(100000)).
		WillReturnRows(sqlmock.NewRows(deploymentColumns).
			AddRow(depID, userID, stratID, "AAPL", int64(100000), int16(0), false, int64(0), int64(0), time.Now(), time.Now()))

	resp, err := h.DeployStrategy(context.Background(), &userpb.DeployStrategyRequest{
		StrategyId: stratID.String(), UserId: userID.String(), Symbol: "aapl", PositionSizeCents: 100000,
	})
	require.NoError(t, err)
	assert.Equal(t, "running", resp.Status)
	assert.Equal(t, "AAPL", resp.Symbol)
}

func TestDeployStrategyReactivatesStopped(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()
	depID := uuid.New()
	cfg := []byte(`{"entry":{"rules":[{"lhs":{"kind":"price"},"op":">","rhs":{"kind":"value","value":1}}]}}`)

	mock.ExpectQuery(`FROM strategies`).
		WithArgs(stratID, userID).
		WillReturnRows(sqlmock.NewRows(strategyColumns).
			AddRow(stratID, userID, "Dip", cfg, time.Now(), time.Now()))
	mock.ExpectQuery(`FROM strategy_deployments`).
		WithArgs(userID, stratID, "AAPL").
		WillReturnRows(sqlmock.NewRows(deploymentColumns).
			AddRow(depID, userID, stratID, "AAPL", int64(50000), int16(1), false, int64(0), int64(0), time.Now(), time.Now()))
	mock.ExpectQuery(`UPDATE strategy_deployments`).
		WithArgs(depID, int64(100000)).
		WillReturnRows(sqlmock.NewRows(deploymentColumns).
			AddRow(depID, userID, stratID, "AAPL", int64(100000), int16(0), false, int64(0), int64(0), time.Now(), time.Now()))

	resp, err := h.DeployStrategy(context.Background(), &userpb.DeployStrategyRequest{
		StrategyId: stratID.String(), UserId: userID.String(), Symbol: "aapl", PositionSizeCents: 100000,
	})
	require.NoError(t, err)
	assert.Equal(t, "running", resp.Status)
	assert.Equal(t, int64(100000), resp.PositionSizeCents)
}

func TestDeployStrategyTooSmall(t *testing.T) {
	h, _ := newStrategyHandler(t)
	_, err := h.DeployStrategy(context.Background(), &userpb.DeployStrategyRequest{
		StrategyId: uuid.New().String(), UserId: uuid.New().String(), Symbol: "AAPL", PositionSizeCents: 50,
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestStopDeployment(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()
	depID := uuid.New()

	mock.ExpectQuery(`UPDATE strategy_deployments`).
		WithArgs(depID, userID).
		WillReturnRows(sqlmock.NewRows(deploymentColumns).
			AddRow(depID, userID, stratID, "AAPL", int64(100000), int16(1), false, int64(0), int64(0), time.Now(), time.Now()))

	resp, err := h.StopDeployment(context.Background(), &userpb.DeploymentSpecifier{
		Id: depID.String(), UserId: userID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, "stopped", resp.Status)
}

func TestDeleteDeploymentStillRunning(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	depID := uuid.New()

	mock.ExpectExec(`DELETE FROM strategy_deployments`).
		WithArgs(depID, userID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	_, err := h.DeleteDeployment(context.Background(), &userpb.DeploymentSpecifier{
		Id: depID.String(), UserId: userID.String(),
	})
	assert.Equal(t, codes.NotFound, status.Code(err))
}

func TestDeleteDeployment(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	depID := uuid.New()

	mock.ExpectExec(`DELETE FROM strategy_deployments`).
		WithArgs(depID, userID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	_, err := h.DeleteDeployment(context.Background(), &userpb.DeploymentSpecifier{
		Id: depID.String(), UserId: userID.String(),
	})
	require.NoError(t, err)
}

func TestGetDeployments(t *testing.T) {
	h, mock := newStrategyHandler(t)
	userID := uuid.New()
	stratID := uuid.New()
	depID := uuid.New()

	cols := append(append([]string{}, deploymentColumns...), "strategy_name")
	mock.ExpectQuery(`FROM strategy_deployments`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows(cols).
			AddRow(depID, userID, stratID, "AAPL", int64(100000), int16(0), true, int64(15000), int64(6), time.Now(), time.Now(), "Dip"))

	resp, err := h.GetDeployments(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	require.Len(t, resp.Deployments, 1)
	assert.Equal(t, "Dip", resp.Deployments[0].StrategyName)
	assert.True(t, resp.Deployments[0].InPosition)
}

func TestRunBacktestValidation(t *testing.T) {
	h, _ := newStrategyHandler(t)

	_, err := h.RunBacktest(context.Background(), &userpb.BacktestRequest{Symbol: ""})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.RunBacktest(context.Background(), &userpb.BacktestRequest{
		Symbol: "AAPL", InitialCapitalCents: 0,
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))

	_, err = h.RunBacktest(context.Background(), &userpb.BacktestRequest{
		Symbol: "AAPL", InitialCapitalCents: 1000, PositionSizeCents: 100, ConfigJson: "{bad",
	})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestRunBacktestNeedsMarketData(t *testing.T) {
	h, _ := newStrategyHandler(t)
	cfg := `{"entry":{"rules":[{"lhs":{"kind":"price"},"op":">","rhs":{"kind":"value","value":1}}]}}`

	_, err := h.RunBacktest(context.Background(), &userpb.BacktestRequest{
		Symbol: "AAPL", InitialCapitalCents: 1000000, PositionSizeCents: 100000,
		ConfigJson: cfg, Timeframe: "DAY",
	})
	assert.Equal(t, codes.Unavailable, status.Code(err))
}

func TestParseBacktestDate(t *testing.T) {
	_, ok := parseBacktestDate("2026-01-02")
	assert.True(t, ok)
	_, ok = parseBacktestDate("")
	assert.False(t, ok)
}
