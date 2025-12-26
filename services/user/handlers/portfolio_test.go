package handlers

import (
	"context"
	"testing"

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
	return NewPortfolioHandler(sdb, zap.NewNop()), mock
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
