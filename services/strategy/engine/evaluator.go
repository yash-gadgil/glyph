package engine

import (
	"fmt"
	"math"
)

func Evaluate(group RuleGroup, bars []Bar) (bool, error) {
	if len(group.Rules) == 0 {
		return false, nil
	}

	results := make([]bool, 0, len(group.Rules))
	for _, rule := range group.Rules {
		ok, err := evaluateRule(rule, bars)
		if err != nil {
			return false, err
		}
		results = append(results, ok)
	}

	if group.Combinator == "OR" {
		for _, r := range results {
			if r {
				return true, nil
			}
		}
		return false, nil
	}
	for _, r := range results {
		if !r {
			return false, nil
		}
	}
	return true, nil
}

func evaluateRule(rule Rule, bars []Bar) (bool, error) {
	if len(bars) < 2 {
		return false, nil
	}

	lhs, err := Series(rule.LHS, bars)
	if err != nil {
		return false, err
	}

	var rhs []float64
	switch rule.RHS.Kind {
	case "indicator":
		if rule.RHS.Indicator == nil {
			return false, fmt.Errorf("rule rhs declared indicator but carries none")
		}
		rhs, err = Series(*rule.RHS.Indicator, bars)
		if err != nil {
			return false, err
		}
	default:
		rhs = make([]float64, len(bars))
		for i := range rhs {
			rhs[i] = rule.RHS.Value
		}
	}

	n := len(bars)
	curL, curR := lhs[n-1], rhs[n-1]
	prevL, prevR := lhs[n-2], rhs[n-2]

	if math.IsNaN(curL) || math.IsNaN(curR) {
		return false, nil
	}

	switch rule.Op {
	case ">":
		return curL > curR, nil
	case "<":
		return curL < curR, nil
	case ">=":
		return curL >= curR, nil
	case "<=":
		return curL <= curR, nil
	case "crosses_above":
		if math.IsNaN(prevL) || math.IsNaN(prevR) {
			return false, nil
		}
		return prevL <= prevR && curL > curR, nil
	case "crosses_below":
		if math.IsNaN(prevL) || math.IsNaN(prevR) {
			return false, nil
		}
		return prevL >= prevR && curL < curR, nil
	}
	return false, fmt.Errorf("unknown operator %q", rule.Op)
}

func MaxLookback(cfg *Config) int {
	max := 30
	scan := func(ref IndicatorRef) {
		need := 0
		switch ref.Kind {
		case "sma", "ema", "rsi", "atr", "stoch_k", "stoch_d", "bbands_upper", "bbands_middle", "bbands_lower":
			need = int(ref.param("period", 20))
		case "macd_line", "macd_signal", "macd_histogram":
			need = int(ref.param("slow", 26)) + int(ref.param("signal", 9))
		}
		if need > max {
			max = need
		}
	}
	for _, group := range []RuleGroup{cfg.Entry, cfg.Exit} {
		for _, rule := range group.Rules {
			scan(rule.LHS)
			if rule.RHS.Indicator != nil {
				scan(*rule.RHS.Indicator)
			}
		}
	}
	return max*2 + 10
}
