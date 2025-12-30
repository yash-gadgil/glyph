package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApplyFillBuyGrowsLotAtCost(t *testing.T) {
	out, err := ApplyFill(PositionState{}, SettledFill{Side: SideBuy, Qty: 10, PriceCents: 5_000})
	require.NoError(t, err)

	assert.Equal(t, int64(10), out.Position.Qty)
	assert.Equal(t, int64(50_000), out.Position.CostBasis)
	assert.Equal(t, int64(0), out.Position.RealizedPnl)
	assert.Equal(t, int64(-50_000), out.CashDeltaCents)
	assert.Equal(t, int64(0), out.RealizedCents)
}

func TestApplyFillBuyAveragesIntoExistingLot(t *testing.T) {
	pos := PositionState{Qty: 10, CostBasis: 50_000}
	out, err := ApplyFill(pos, SettledFill{Side: SideBuy, Qty: 5, PriceCents: 6_200})
	require.NoError(t, err)

	assert.Equal(t, int64(15), out.Position.Qty)
	assert.Equal(t, int64(81_000), out.Position.CostBasis)
	assert.Equal(t, int64(-31_000), out.CashDeltaCents)
}

func TestApplyFillSellAllRealizesFullDifference(t *testing.T) {
	pos := PositionState{Qty: 10, CostBasis: 50_000}
	out, err := ApplyFill(pos, SettledFill{Side: SideSell, Qty: 10, PriceCents: 6_000})
	require.NoError(t, err)

	assert.Equal(t, int64(0), out.Position.Qty)
	assert.Equal(t, int64(0), out.Position.CostBasis)
	assert.Equal(t, int64(10_000), out.Position.RealizedPnl)
	assert.Equal(t, int64(60_000), out.CashDeltaCents)
}

func TestApplyFillPartialSellRelievesCostProportionally(t *testing.T) {
	pos := PositionState{Qty: 10, CostBasis: 50_000}
	out, err := ApplyFill(pos, SettledFill{Side: SideSell, Qty: 4, PriceCents: 6_000})
	require.NoError(t, err)

	assert.Equal(t, int64(6), out.Position.Qty)
	assert.Equal(t, int64(30_000), out.Position.CostBasis)
	assert.Equal(t, int64(4_000), out.Position.RealizedPnl)
	assert.Equal(t, int64(24_000), out.CashDeltaCents)
}

func TestApplyFillSellRoundsHalfToEven(t *testing.T) {
	pos := PositionState{Qty: 3, CostBasis: 100}
	out, err := ApplyFill(pos, SettledFill{Side: SideSell, Qty: 1, PriceCents: 50})
	require.NoError(t, err)

	assert.Equal(t, int64(2), out.Position.Qty)
	assert.Equal(t, int64(67), out.Position.CostBasis)
}

func TestApplyFillRejectsBadInput(t *testing.T) {
	_, err := ApplyFill(PositionState{}, SettledFill{Side: SideBuy, Qty: 0, PriceCents: 100})
	assert.Error(t, err)

	_, err = ApplyFill(PositionState{Qty: 1, CostBasis: 100}, SettledFill{Side: SideSell, Qty: 5, PriceCents: 100})
	assert.Error(t, err)

	_, err = ApplyFill(PositionState{}, SettledFill{Side: 9, Qty: 1, PriceCents: 100})
	assert.Error(t, err)
}
