package worker

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/user/types"
	"go.uber.org/zap"
)

func newSettler(t *testing.T) (*Settler, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	return NewSettler(sdb, zap.NewNop()), mock
}

func fillEvent(side int16, qty, price int64) types.FillEvent {
	return types.FillEvent{
		TradeID:        uuid.NewString(),
		Symbol:         "AAPL",
		OrderID:        uuid.NewString(),
		CounterOrderID: uuid.NewString(),
		UserID:         uuid.NewString(),
		Side:           side,
		Qty:            qty,
		PriceCents:     price,
		Liquidity:      "taker",
		ExecutedAt:     time.Now(),
	}
}

var reservationColumns = []string{"order_id", "user_id", "symbol", "side", "qty", "remaining_qty", "cents_per_share", "created_at"}
var positionLockColumns = []string{"symbol", "qty", "reserved_qty", "realized_pnl", "cost_basis"}

func TestSettleFillBuySettlesCashAndPosition(t *testing.T) {
	settler, mock := newSettler(t)
	ev := fillEvent(0, 10, 10_000)
	orderID := uuid.MustParse(ev.OrderID)
	userID := uuid.MustParse(ev.UserID)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM positions\s+WHERE user_id = \$1 AND symbol = \$2\s+FOR UPDATE`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO settlements`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO positions`).
		WithArgs(userID, "AAPL", int64(10), int64(0), int64(100_000)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`FROM order_reservations\s+WHERE order_id = \$1\s+FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows(reservationColumns).
			AddRow(orderID, userID, "AAPL", int16(0), int64(10), int64(10), int64(10_000), time.Now()))
	mock.ExpectQuery(`UPDATE order_reservations\s+SET remaining_qty = remaining_qty -`).
		WithArgs(orderID, int64(10)).
		WillReturnRows(sqlmock.NewRows(reservationColumns).
			AddRow(orderID, userID, "AAPL", int16(0), int64(10), int64(0), int64(10_000), time.Now()))
	mock.ExpectExec(`UPDATE accounts\s+SET cash_balance = cash_balance \+`).
		WithArgs(userID, int64(-100_000), int64(100_000)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, settler.SettleFill(context.Background(), ev))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSettleFillSellRealizesPnlAndReleasesShares(t *testing.T) {
	settler, mock := newSettler(t)
	ev := fillEvent(1, 10, 6_000)
	orderID := uuid.MustParse(ev.OrderID)
	userID := uuid.MustParse(ev.UserID)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM positions\s+WHERE user_id = \$1 AND symbol = \$2\s+FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows(positionLockColumns).
			AddRow("AAPL", int64(10), int64(10), int64(0), int64(50_000)))
	mock.ExpectExec(`INSERT INTO settlements`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO positions`).
		WithArgs(userID, "AAPL", int64(0), int64(10_000), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`FROM order_reservations\s+WHERE order_id = \$1\s+FOR UPDATE`).
		WillReturnRows(sqlmock.NewRows(reservationColumns).
			AddRow(orderID, userID, "AAPL", int16(1), int64(10), int64(10), int64(0), time.Now()))
	mock.ExpectQuery(`UPDATE order_reservations\s+SET remaining_qty = remaining_qty -`).
		WillReturnRows(sqlmock.NewRows(reservationColumns).
			AddRow(orderID, userID, "AAPL", int16(1), int64(10), int64(0), int64(0), time.Now()))
	mock.ExpectExec(`UPDATE accounts\s+SET cash_balance = cash_balance \+`).
		WithArgs(userID, int64(60_000), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE positions\s+SET reserved_qty = GREATEST`).
		WithArgs(userID, "AAPL", int64(10)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, settler.SettleFill(context.Background(), ev))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSettleFillDuplicateDeliveryIsNoop(t *testing.T) {
	settler, mock := newSettler(t)
	ev := fillEvent(0, 10, 10_000)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM positions`).
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO settlements`).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	require.NoError(t, settler.SettleFill(context.Background(), ev))
	assert.NoError(t, mock.ExpectationsWereMet(), "no position/cash writes on redelivery")
}

func TestSettleFillWithoutReservationStillSettlesCash(t *testing.T) {
	settler, mock := newSettler(t)
	ev := fillEvent(0, 5, 2_000)
	userID := uuid.MustParse(ev.UserID)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM positions`).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`INSERT INTO settlements`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`INSERT INTO positions`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`FROM order_reservations`).WillReturnError(sql.ErrNoRows)
	mock.ExpectExec(`UPDATE accounts\s+SET cash_balance = cash_balance \+`).
		WithArgs(userID, int64(-10_000), int64(0)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, settler.SettleFill(context.Background(), ev))
}

func TestSettleFillRejectsMalformedEvents(t *testing.T) {
	settler, _ := newSettler(t)
	ctx := context.Background()

	bad := fillEvent(0, 10, 100)
	bad.TradeID = "nope"
	assert.Error(t, settler.SettleFill(ctx, bad))

	bad = fillEvent(0, 10, 100)
	bad.UserID = "nope"
	assert.Error(t, settler.SettleFill(ctx, bad))

	bad = fillEvent(0, 0, 100)
	assert.Error(t, settler.SettleFill(ctx, bad))
}

func TestHandleDoneReleasesBuyRemainder(t *testing.T) {
	settler, mock := newSettler(t)
	orderID := uuid.New()
	userID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM order_reservations\s+WHERE order_id = \$1\s+FOR UPDATE`).
		WithArgs(orderID).
		WillReturnRows(sqlmock.NewRows(reservationColumns).
			AddRow(orderID, userID, "AAPL", int16(0), int64(10), int64(6), int64(10_000), time.Now()))
	mock.ExpectExec(`UPDATE accounts\s+SET reserved_cash = GREATEST`).
		WithArgs(userID, int64(60_000)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`DELETE FROM order_reservations`).
		WithArgs(orderID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	require.NoError(t, settler.HandleDone(context.Background(), types.DoneEvent{
		OrderID: orderID.String(),
		Reason:  "ioc_expired",
	}))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestHandleDoneUnknownReservationIsNoop(t *testing.T) {
	settler, mock := newSettler(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`FROM order_reservations`).WillReturnError(sql.ErrNoRows)
	mock.ExpectCommit()

	require.NoError(t, settler.HandleDone(context.Background(), types.DoneEvent{
		OrderID: uuid.NewString(),
		Reason:  "cancelled",
	}))
}
