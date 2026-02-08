package strategyengine

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func priceStrategy(entryAbove, exitAbove, sl, tp float64) *Config {
	return &Config{
		Name: "test",
		Entry: RuleGroup{Combinator: "AND", Rules: []Rule{{
			LHS: IndicatorRef{Kind: "price"}, Op: ">", RHS: RuleRHS{Kind: "value", Value: entryAbove},
		}}},
		Exit: RuleGroup{Combinator: "AND", Rules: []Rule{{
			LHS: IndicatorRef{Kind: "price"}, Op: ">", RHS: RuleRHS{Kind: "value", Value: exitAbove},
		}}},
		StopLossPct:   sl,
		TakeProfitPct: tp,
	}
}

type ohlc struct{ o, h, l, c float64 }

func barsWithWarmup(flat float64, warm int, extra ...ohlc) []Bar {
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := make([]Bar, 0, warm+len(extra))
	for i := 0; i < warm; i++ {
		bars = append(bars, Bar{Time: base.AddDate(0, 0, len(bars)), Open: flat, High: flat, Low: flat, Close: flat, Volume: 1000})
	}
	for _, e := range extra {
		bars = append(bars, Bar{Time: base.AddDate(0, 0, len(bars)), Open: e.o, High: e.h, Low: e.l, Close: e.c, Volume: 1000})
	}
	return bars
}

var defaultParams = BacktestParams{InitialCapitalCents: 10_000_000, PositionSizeCents: 1_000_000, Timeframe: "DAY"}

func TestBacktestNextBarOpenFillAndMetrics(t *testing.T) {
	cfg := priceStrategy(100, 110, 0, 0)
	bars := barsWithWarmup(100, 71,
		ohlc{105, 106, 104, 105},
		ohlc{106, 109, 104, 108},
		ohlc{108, 113, 107, 112},
		ohlc{111, 112, 110, 111},
	)
	require.Equal(t, 75, len(bars))

	res, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)

	require.Len(t, res.Trades, 1)
	tr := res.Trades[0]
	assert.Equal(t, int64(10600), tr.EntryPriceCents)
	assert.Equal(t, int64(11100), tr.ExitPriceCents)
	assert.Equal(t, 72, tr.EntryIndex)
	assert.Equal(t, 74, tr.ExitIndex)
	assert.Equal(t, 2, tr.HoldBars)
	assert.Equal(t, "signal", tr.ExitReason)

	assert.Equal(t, int64(94), tr.Qty)
	assert.Equal(t, int64(47_000), tr.PnLCents)

	assert.Equal(t, int64(10_047_000), res.FinalEquityCents)
	assert.InDelta(t, 0.47, res.TotalReturnPct, 1e-9)
	assert.Equal(t, 1, res.NumTrades)
	assert.InDelta(t, 100.0, res.WinRate, 1e-9)
	assert.InDelta(t, 2.0, res.AvgHoldBars, 1e-9)
	assert.Equal(t, 0.0, res.ProfitFactor)
	assert.Equal(t, 70, res.WarmupBars)
	assert.Equal(t, 75, res.BarsUsed)
}

func TestBacktestStopLossAtLevel(t *testing.T) {
	cfg := priceStrategy(100, 99999, 2, 0)
	bars := barsWithWarmup(100, 71,
		ohlc{105, 106, 104, 105},
		ohlc{100, 101, 99, 100},
		ohlc{99, 100, 97, 98},
	)
	res, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)

	require.Len(t, res.Trades, 1)
	tr := res.Trades[0]
	assert.Equal(t, int64(10000), tr.EntryPriceCents)
	assert.Equal(t, int64(9800), tr.ExitPriceCents)
	assert.Equal(t, "stop_loss", tr.ExitReason)
	assert.InDelta(t, 0.0, res.WinRate, 1e-9)
}

func TestBacktestStopLossGapsThroughAtOpen(t *testing.T) {
	cfg := priceStrategy(100, 99999, 2, 0)
	bars := barsWithWarmup(100, 71,
		ohlc{105, 106, 104, 105},
		ohlc{100, 101, 99, 100},
		ohlc{96, 97, 95, 96},
	)
	res, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)

	require.Len(t, res.Trades, 1)
	assert.Equal(t, int64(9600), res.Trades[0].ExitPriceCents)
	assert.Equal(t, "stop_loss", res.Trades[0].ExitReason)
}

func TestBacktestTakeProfitAtLevel(t *testing.T) {
	cfg := priceStrategy(100, 99999, 0, 4)
	bars := barsWithWarmup(100, 71,
		ohlc{105, 106, 104, 105},
		ohlc{100, 101, 99, 100},
		ohlc{101, 105, 100, 104},
	)
	res, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)

	require.Len(t, res.Trades, 1)
	assert.Equal(t, int64(10400), res.Trades[0].ExitPriceCents)
	assert.Equal(t, "take_profit", res.Trades[0].ExitReason)
}

func TestBacktestForceClosesAtEndOfData(t *testing.T) {
	cfg := priceStrategy(100, 99999, 0, 0)
	bars := barsWithWarmup(100, 71,
		ohlc{105, 106, 104, 105},
		ohlc{106, 107, 105, 106},
		ohlc{106, 108, 105, 107},
		ohlc{107, 109, 106, 108},
	)
	res, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)

	require.Len(t, res.Trades, 1)
	tr := res.Trades[0]
	assert.Equal(t, "end_of_data", tr.ExitReason)
	assert.Equal(t, int64(10800), tr.ExitPriceCents)
	assert.Equal(t, 74, tr.ExitIndex)
}

func TestBacktestIgnoresPreWarmupSignals(t *testing.T) {
	cfg := priceStrategy(100, 110, 0, 0)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	bars := make([]Bar, 0, 75)
	for i := 0; i < 70; i++ {
		bars = append(bars, Bar{Time: base.AddDate(0, 0, i), Open: 105, High: 106, Low: 104, Close: 105, Volume: 1000})
	}
	for i := 70; i < 75; i++ {
		bars = append(bars, Bar{Time: base.AddDate(0, 0, i), Open: 95, High: 96, Low: 94, Close: 95, Volume: 1000})
	}

	res, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)
	assert.Equal(t, 0, res.NumTrades)
	assert.Equal(t, 70, res.WarmupBars)
	assert.Equal(t, 75, res.BarsUsed)
	assert.InDelta(t, 0.0, res.TotalReturnPct, 1e-9)
}

func TestBacktestRejectsTooFewBars(t *testing.T) {
	cfg := priceStrategy(100, 110, 0, 0)
	bars := barsWithWarmup(100, 70)
	_, err := RunBacktest(cfg, bars, defaultParams)
	assert.Error(t, err)
}

func TestBacktestDoesNotMutateInput(t *testing.T) {
	cfg := priceStrategy(100, 110, 2, 4)
	bars := barsWithWarmup(100, 71,
		ohlc{105, 106, 104, 105},
		ohlc{106, 109, 104, 108},
		ohlc{108, 113, 107, 112},
		ohlc{111, 112, 110, 111},
	)
	snapshot := make([]Bar, len(bars))
	copy(snapshot, bars)

	_, err := RunBacktest(cfg, bars, defaultParams)
	require.NoError(t, err)
	assert.True(t, reflect.DeepEqual(snapshot, bars), "RunBacktest must not mutate input bars")
}
