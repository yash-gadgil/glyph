package handlers

import (
	"context"
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

func newPortfolioHandler(t *testing.T) (*PortfolioHandler, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	return NewPortfolioHandler(sdb, nil, zap.NewNop()), mock
}

func TestGetPortfolio(t *testing.T) {
	h, mock := newPortfolioHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`FROM accounts`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"cash_balance", "reserved_cash", "currency", "multiplier", "margin_used"}).
			AddRow(int64(10000000), int64(2000), "USD", int32(1), int64(0)))

	resp, err := h.GetPortfolio(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	assert.Equal(t, int64(10000000), resp.CashBalanceCents)
	assert.Equal(t, "USD", resp.Currency)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPortfolioInvalidUser(t *testing.T) {
	h, _ := newPortfolioHandler(t)

	_, err := h.GetPortfolio(context.Background(), &userpb.UserSpecifier{UserId: "bad"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetHoldingsValuesAtCostBasis(t *testing.T) {
	h, mock := newPortfolioHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`FROM positions`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "qty", "reserved_qty", "realized_pnl", "cost_basis", "updated_at"}).
			AddRow("AAPL", int64(10), int64(0), int64(100), int64(50000), time.Now()))

	resp, err := h.GetHoldings(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	require.Len(t, resp.Holdings, 1)
	assert.Equal(t, int64(5000), resp.Holdings[0].AvgPriceCents)
	assert.Equal(t, int64(50000), resp.Holdings[0].MarketValueCents)
	assert.Equal(t, int64(0), resp.TotalUnrealizedPnlCents)
	assert.Equal(t, int64(50000), resp.TotalCostBasisCents)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHoldingsSkipsClosedPositions(t *testing.T) {
	h, mock := newPortfolioHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`FROM positions`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "qty", "reserved_qty", "realized_pnl", "cost_basis", "updated_at"}).
			AddRow("AAPL", int64(0), int64(0), int64(0), int64(0), time.Now()))

	resp, err := h.GetHoldings(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	assert.Empty(t, resp.Holdings)
}

func TestGetPositions(t *testing.T) {
	h, mock := newPortfolioHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`FROM positions`).
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"symbol", "qty", "reserved_qty", "realized_pnl", "cost_basis", "updated_at"}).
			AddRow("AAPL", int64(10), int64(2), int64(0), int64(50000), time.Now()))

	resp, err := h.GetPositions(context.Background(), &userpb.UserSpecifier{UserId: userID.String()})
	require.NoError(t, err)
	require.Len(t, resp.Positions, 1)
	assert.Equal(t, int64(2), resp.Positions[0].ReservedQty)
	assert.Equal(t, int64(5000), resp.Positions[0].AvgPriceCents)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetPositionsInvalidUser(t *testing.T) {
	h, _ := newPortfolioHandler(t)

	_, err := h.GetPositions(context.Background(), &userpb.UserSpecifier{UserId: "bad"})
	assert.Equal(t, codes.InvalidArgument, status.Code(err))
}

func TestGetPortfolioHistory(t *testing.T) {
	h, mock := newPortfolioHandler(t)
	userID := uuid.New()

	mock.ExpectQuery(`FROM account_value_snapshots`).
		WithArgs(userID, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"user_id", "equity_cents", "cash_cents", "market_value_cents", "captured_at"}).
			AddRow(userID, int64(10050000), int64(9000000), int64(1050000), time.Now()))

	resp, err := h.GetPortfolioHistory(context.Background(), &userpb.PortfolioHistoryRequest{
		UserId: userID.String(), Hours: 48,
	})
	require.NoError(t, err)
	require.Len(t, resp.Points, 1)
	assert.Equal(t, int64(10050000), resp.Points[0].EquityCents)
}
