package worker

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type fakePriceSource struct {
	mrktpb.MrktdataServiceClient
	prices map[string]int64
	err    error
	calls  [][]string
}

func (f *fakePriceSource) GetLatestPrices(_ context.Context, req *mrktpb.LatestPricesRequest, _ ...grpc.CallOption) (*mrktpb.LatestPricesResponse, error) {
	f.calls = append(f.calls, req.Symbols)
	if f.err != nil {
		return nil, f.err
	}
	resp := &mrktpb.LatestPricesResponse{}
	for sym, cents := range f.prices {
		resp.Prices = append(resp.Prices, &mrktpb.SymbolPrice{Symbol: sym, PriceCents: cents})
	}
	return resp, nil
}

func newSnapshotter(t *testing.T, prices *fakePriceSource) (*Snapshotter, sqlmock.Sqlmock) {
	t.Helper()
	sdb, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	t.Cleanup(func() { _ = sdb.Close() })
	if prices == nil {
		return NewSnapshotter(sdb, nil, zap.NewNop()), mock
	}
	return NewSnapshotter(sdb, prices, zap.NewNop()), mock
}

var (
	accountSnapColumns  = []string{"user_id", "cash_balance"}
	positionSnapColumns = []string{"user_id", "symbol", "qty", "cost_basis"}
)

func TestSnapshotOnceValuesEquityWithLivePrices(t *testing.T) {
	prices := &fakePriceSource{prices: map[string]int64{"AAPL": 20_000, "TSLA": 25_000}}
	snap, mock := newSnapshotter(t, prices)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT user_id, cash_balance\s+FROM accounts`).
		WillReturnRows(sqlmock.NewRows(accountSnapColumns).
			AddRow(userID, int64(1_000_000)))
	mock.ExpectQuery(`SELECT user_id, symbol, qty, cost_basis\s+FROM positions\s+WHERE qty != 0`).
		WillReturnRows(sqlmock.NewRows(positionSnapColumns).
			AddRow(userID, "AAPL", int64(10), int64(180_000)).
			AddRow(userID, "TSLA", int64(2), int64(48_000)))

	mock.ExpectExec(`INSERT INTO account_value_snapshots`).
		WithArgs(userID, int64(1_000_000+250_000), int64(1_000_000), int64(250_000)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, snap.snapshotOnce(context.Background()))
	assert.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, prices.calls, 1)
	assert.ElementsMatch(t, []string{"AAPL", "TSLA"}, prices.calls[0])
}

func TestSnapshotOnceFallsBackToCostBasisWhenPriceMissing(t *testing.T) {
	prices := &fakePriceSource{prices: map[string]int64{"AAPL": 20_000}}
	snap, mock := newSnapshotter(t, prices)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT user_id, cash_balance\s+FROM accounts`).
		WillReturnRows(sqlmock.NewRows(accountSnapColumns).
			AddRow(userID, int64(500_000)))
	mock.ExpectQuery(`SELECT user_id, symbol, qty, cost_basis\s+FROM positions`).
		WillReturnRows(sqlmock.NewRows(positionSnapColumns).
			AddRow(userID, "AAPL", int64(10), int64(180_000)).
			AddRow(userID, "TSLA", int64(4), int64(96_000)))

	mock.ExpectExec(`INSERT INTO account_value_snapshots`).
		WithArgs(userID, int64(500_000+296_000), int64(500_000), int64(296_000)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, snap.snapshotOnce(context.Background()))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSnapshotOnceCostBasisWhenNoPriceSource(t *testing.T) {
	snap, mock := newSnapshotter(t, nil)
	userID := uuid.New()

	mock.ExpectQuery(`SELECT user_id, cash_balance\s+FROM accounts`).
		WillReturnRows(sqlmock.NewRows(accountSnapColumns).
			AddRow(userID, int64(700_000)))
	mock.ExpectQuery(`SELECT user_id, symbol, qty, cost_basis\s+FROM positions`).
		WillReturnRows(sqlmock.NewRows(positionSnapColumns).
			AddRow(userID, "AAPL", int64(10), int64(180_000)))

	mock.ExpectExec(`INSERT INTO account_value_snapshots`).
		WithArgs(userID, int64(700_000+180_000), int64(700_000), int64(180_000)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, snap.snapshotOnce(context.Background()))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSnapshotOnceNoAccountsIsNoop(t *testing.T) {
	snap, mock := newSnapshotter(t, &fakePriceSource{})

	mock.ExpectQuery(`SELECT user_id, cash_balance\s+FROM accounts`).
		WillReturnRows(sqlmock.NewRows(accountSnapColumns))

	require.NoError(t, snap.snapshotOnce(context.Background()))
	assert.NoError(t, mock.ExpectationsWereMet())
}
