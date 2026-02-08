package strategyengine

import (
	"fmt"
	"math"
	"time"
)

type BacktestParams struct {
	InitialCapitalCents int64
	PositionSizeCents   int64
	Timeframe           string
}

type BacktestTrade struct {
	EntryIndex      int
	ExitIndex       int
	EntryTime       time.Time
	ExitTime        time.Time
	EntryPriceCents int64
	ExitPriceCents  int64
	Qty             int64
	PnLCents        int64
	ReturnPct       float64
	HoldBars        int
	ExitReason      string
}

type EquityPoint struct {
	Time        time.Time
	EquityCents int64
}

type BacktestResult struct {
	TotalReturnPct   float64
	MaxDrawdownPct   float64
	Sharpe           float64
	WinRate          float64
	ProfitFactor     float64
	NumTrades        int
	AvgHoldBars      float64
	FinalEquityCents int64
	EquityCurve      []EquityPoint
	Trades           []BacktestTrade
	BarsUsed         int
	WarmupBars       int
}

func cents(d float64) int64 { return int64(math.Round(d * 100)) }

func RunBacktest(cfg *Config, bars []Bar, p BacktestParams) (*BacktestResult, error) {
	if cfg == nil {
		return nil, fmt.Errorf("backtest: nil config")
	}
	if p.InitialCapitalCents <= 0 {
		return nil, fmt.Errorf("backtest: initial capital must be positive")
	}
	if p.PositionSizeCents <= 0 {
		return nil, fmt.Errorf("backtest: position size must be positive")
	}

	warmup := MaxLookback(cfg)
	if len(bars) <= warmup {
		return nil, fmt.Errorf("backtest: need more than %d bars (warm-up), got %d", warmup, len(bars))
	}

	n := len(bars)
	cash := p.InitialCapitalCents

	var (
		inPosition   bool
		entryCents   int64
		entryPrice   float64
		entryIndex   int
		entryTime    time.Time
		qty          int64
		pendingEntry bool
		pendingExit  bool

		trades      []BacktestTrade
		equityCurve = make([]EquityPoint, 0, n-warmup)
	)

	closeTrade := func(i int, exitPrice float64, reason string) {
		exitCents := cents(exitPrice)
		cash += qty * exitCents
		pnl := qty * (exitCents - entryCents)
		var ret float64
		if entryCents > 0 {
			ret = float64(exitCents-entryCents) / float64(entryCents) * 100
		}
		trades = append(trades, BacktestTrade{
			EntryIndex:      entryIndex,
			ExitIndex:       i,
			EntryTime:       entryTime,
			ExitTime:        bars[i].Time,
			EntryPriceCents: entryCents,
			ExitPriceCents:  exitCents,
			Qty:             qty,
			PnLCents:        pnl,
			ReturnPct:       ret,
			HoldBars:        i - entryIndex,
			ExitReason:      reason,
		})
		inPosition = false
		qty = 0
	}

	for i := warmup; i < n; i++ {
		bar := bars[i]

		if pendingExit && inPosition {
			closeTrade(i, bar.Open, "signal")
		}
		pendingExit = false
		if pendingEntry && !inPosition {
			ec := cents(bar.Open)
			budget := p.PositionSizeCents
			if cash < budget {
				budget = cash
			}
			if ec > 0 {
				if q := budget / ec; q >= 1 {
					qty = q
					entryCents = ec
					entryPrice = bar.Open
					entryIndex = i
					entryTime = bar.Time
					cash -= qty * entryCents
					inPosition = true
				}
			}
		}
		pendingEntry = false

		if inPosition && cfg.StopLossPct > 0 {
			slLevel := entryPrice * (1 - cfg.StopLossPct/100)
			if bar.Low <= slLevel {
				fill := slLevel
				if bar.Open <= slLevel {
					fill = bar.Open
				}
				closeTrade(i, fill, "stop_loss")
			}
		}
		if inPosition && cfg.TakeProfitPct > 0 {
			tpLevel := entryPrice * (1 + cfg.TakeProfitPct/100)
			if bar.High >= tpLevel {
				fill := tpLevel
				if bar.Open >= tpLevel {
					fill = bar.Open
				}
				closeTrade(i, fill, "take_profit")
			}
		}

		window := bars[:i+1]
		if inPosition {
			exit, err := Evaluate(cfg.Exit, window)
			if err != nil {
				return nil, fmt.Errorf("backtest: exit rules: %w", err)
			}
			if exit {
				pendingExit = true
			}
		} else {
			enter, err := Evaluate(cfg.Entry, window)
			if err != nil {
				return nil, fmt.Errorf("backtest: entry rules: %w", err)
			}
			if enter {
				pendingEntry = true
			}
		}

		equity := cash
		if inPosition {
			equity += qty * cents(bar.Close)
		}
		equityCurve = append(equityCurve, EquityPoint{Time: bar.Time, EquityCents: equity})
	}

	if inPosition {
		last := n - 1
		closeTrade(last, bars[last].Close, "end_of_data")
		if len(equityCurve) > 0 {
			equityCurve[len(equityCurve)-1].EquityCents = cash
		}
	}

	result := &BacktestResult{
		FinalEquityCents: cash,
		EquityCurve:      equityCurve,
		Trades:           trades,
		NumTrades:        len(trades),
		BarsUsed:         n,
		WarmupBars:       warmup,
	}
	result.TotalReturnPct = float64(cash-p.InitialCapitalCents) / float64(p.InitialCapitalCents) * 100

	peak := p.InitialCapitalCents
	for _, pt := range equityCurve {
		if pt.EquityCents > peak {
			peak = pt.EquityCents
		}
		if peak > 0 {
			if dd := float64(peak-pt.EquityCents) / float64(peak) * 100; dd > result.MaxDrawdownPct {
				result.MaxDrawdownPct = dd
			}
		}
	}

	result.Sharpe = sharpe(equityCurve, periodsPerYear(p.Timeframe))

	var wins int
	var grossProfit, grossLoss, holdSum float64
	for _, t := range trades {
		switch {
		case t.PnLCents > 0:
			wins++
			grossProfit += float64(t.PnLCents)
		case t.PnLCents < 0:
			grossLoss += float64(-t.PnLCents)
		}
		holdSum += float64(t.HoldBars)
	}
	if len(trades) > 0 {
		result.WinRate = float64(wins) / float64(len(trades)) * 100
		result.AvgHoldBars = holdSum / float64(len(trades))
	}
	if grossLoss > 0 {
		result.ProfitFactor = grossProfit / grossLoss
	}

	return result, nil
}

func periodsPerYear(tf string) float64 {
	switch tf {
	case "MIN":
		return 252 * 390
	case "HOUR":
		return 252 * 6.5
	default:
		return 252
	}
}

func sharpe(curve []EquityPoint, ppy float64) float64 {
	if len(curve) < 3 {
		return 0
	}
	rets := make([]float64, 0, len(curve)-1)
	for i := 1; i < len(curve); i++ {
		prev := curve[i-1].EquityCents
		if prev == 0 {
			continue
		}
		rets = append(rets, float64(curve[i].EquityCents-prev)/float64(prev))
	}
	if len(rets) < 2 {
		return 0
	}

	var mean float64
	for _, r := range rets {
		mean += r
	}
	mean /= float64(len(rets))

	var variance float64
	for _, r := range rets {
		d := r - mean
		variance += d * d
	}
	variance /= float64(len(rets) - 1)
	std := math.Sqrt(variance)
	if std == 0 {
		return 0
	}
	return mean / std * math.Sqrt(ppy)
}
