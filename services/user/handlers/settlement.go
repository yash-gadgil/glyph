package handlers

import (
	"fmt"
	"math/big"
)

type PositionState struct {
	Qty         int64
	CostBasis   int64
	RealizedPnl int64
}

type SettledFill struct {
	Side       int16
	Qty        int64
	PriceCents int64
}

type FillOutcome struct {
	Position       PositionState
	CashDeltaCents int64
	RealizedCents  int64
}

func ApplyFill(pos PositionState, fill SettledFill) (FillOutcome, error) {
	if fill.Qty <= 0 {
		return FillOutcome{}, fmt.Errorf("fill qty must be positive, got %d", fill.Qty)
	}
	if fill.PriceCents < 0 {
		return FillOutcome{}, fmt.Errorf("fill price must be non-negative, got %d", fill.PriceCents)
	}

	switch fill.Side {
	case SideBuy:
		pos.Qty += fill.Qty
		pos.CostBasis += fill.Qty * fill.PriceCents
		return FillOutcome{
			Position:       pos,
			CashDeltaCents: -fill.Qty * fill.PriceCents,
		}, nil

	case SideSell:
		if fill.Qty > pos.Qty {
			return FillOutcome{}, fmt.Errorf("sell qty %d exceeds position qty %d", fill.Qty, pos.Qty)
		}

		costRemoved := roundHalfEvenDiv(pos.CostBasis, fill.Qty, pos.Qty)
		proceeds := fill.Qty * fill.PriceCents
		realized := proceeds - costRemoved

		pos.Qty -= fill.Qty
		pos.CostBasis -= costRemoved
		pos.RealizedPnl += realized

		if pos.Qty == 0 && pos.CostBasis != 0 {
			return FillOutcome{}, fmt.Errorf("invariant violation: empty position retains cost basis %d", pos.CostBasis)
		}

		return FillOutcome{
			Position:       pos,
			CashDeltaCents: proceeds,
			RealizedCents:  realized,
		}, nil

	default:
		return FillOutcome{}, fmt.Errorf("unknown side %d", fill.Side)
	}
}

func roundHalfEvenDiv(a, b, den int64) int64 {
	num := new(big.Int).Mul(big.NewInt(a), big.NewInt(b))
	d := big.NewInt(den)

	q, r := new(big.Int).QuoRem(num, d, new(big.Int))

	twice := new(big.Int).Lsh(r, 1)
	switch twice.Cmp(d) {
	case 1:
		q.Add(q, big.NewInt(1))
	case 0:
		if q.Bit(0) == 1 {
			q.Add(q, big.NewInt(1))
		}
	}

	return q.Int64()
}
