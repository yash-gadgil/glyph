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
