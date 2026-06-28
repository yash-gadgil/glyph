package engine

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func closesToBars(closes ...float64) []Bar {
	bars := make([]Bar, len(closes))
	for i, c := range closes {
		bars[i] = Bar{Open: c, High: c + 1, Low: c - 1, Close: c, Volume: 1000}
	}
	return bars
}

func ref(kind string, params map[string]float64) IndicatorRef {
	return IndicatorRef{Kind: kind, Params: params}
}

func TestSMAWarmupAndValues(t *testing.T) {
	out, err := Series(ref("sma", map[string]float64{"period": 3}), closesToBars(1, 2, 3, 4, 5))
	require.NoError(t, err)
	assert.True(t, math.IsNaN(out[0]))
	assert.True(t, math.IsNaN(out[1]))
	assert.InDelta(t, 2.0, out[2], 1e-9)
	assert.InDelta(t, 3.0, out[3], 1e-9)
	assert.InDelta(t, 4.0, out[4], 1e-9)
}

func TestEMAConvergesTowardRecentPrices(t *testing.T) {
	closes := make([]float64, 50)
	for i := range closes {
		closes[i] = 100
	}
	closes[49] = 110
	out, err := Series(ref("ema", map[string]float64{"period": 10}), closesToBars(closes...))
	require.NoError(t, err)
	last := out[len(out)-1]
	assert.Greater(t, last, 100.0)
	assert.Less(t, last, 110.0)
}

func TestRSIExtremes(t *testing.T) {
	rising := make([]float64, 30)
	for i := range rising {
		rising[i] = 100 + float64(i)
	}
	out, err := Series(ref("rsi", map[string]float64{"period": 14}), closesToBars(rising...))
	require.NoError(t, err)
	assert.InDelta(t, 100.0, out[len(out)-1], 1e-9)

	falling := make([]float64, 30)
	for i := range falling {
		falling[i] = 200 - float64(i)
	}
	out, err = Series(ref("rsi", map[string]float64{"period": 14}), closesToBars(falling...))
	require.NoError(t, err)
	assert.InDelta(t, 0.0, out[len(out)-1], 1e-9)
}

func TestBollingerBandsSurroundTheMean(t *testing.T) {
	bars := closesToBars(10, 12, 11, 13, 12, 14, 13, 15, 14, 16,
		15, 17, 16, 18, 17, 19, 18, 20, 19, 21, 20)
	upper, err := Series(ref("bbands_upper", map[string]float64{"period": 20, "stddev": 2}), bars)
	require.NoError(t, err)
	mid, err := Series(ref("bbands_middle", map[string]float64{"period": 20}), bars)
	require.NoError(t, err)
	lower, err := Series(ref("bbands_lower", map[string]float64{"period": 20, "stddev": 2}), bars)
	require.NoError(t, err)

	n := len(bars) - 1
	assert.Greater(t, upper[n], mid[n])
	assert.Less(t, lower[n], mid[n])
	assert.InDelta(t, mid[n]-lower[n], upper[n]-mid[n], 1e-9, "bands are symmetric")
}

func TestStochKBounds(t *testing.T) {
	bars := closesToBars(10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24)
	out, err := Series(ref("stoch_k", map[string]float64{"period": 14}), bars)
	require.NoError(t, err)
	last := out[len(out)-1]
	assert.GreaterOrEqual(t, last, 0.0)
	assert.LessOrEqual(t, last, 100.0)
	assert.Greater(t, last, 80.0)
}

func TestUnknownIndicatorErrors(t *testing.T) {
	_, err := Series(ref("astrology", nil), closesToBars(1, 2, 3))
	assert.Error(t, err)
}

func TestEvaluateThresholdRule(t *testing.T) {
	group := RuleGroup{
		Combinator: "AND",
		Rules: []Rule{{
			LHS: ref("price", nil),
			Op:  ">",
			RHS: RuleRHS{Kind: "value", Value: 100},
		}},
	}
	ok, err := Evaluate(group, closesToBars(90, 95, 105))
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Evaluate(group, closesToBars(90, 95, 99))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestEvaluateCrossesAbove(t *testing.T) {
	group := RuleGroup{
		Combinator: "AND",
		Rules: []Rule{{
			LHS: ref("price", nil),
			Op:  "crosses_above",
			RHS: RuleRHS{Kind: "value", Value: 100},
		}},
	}

	ok, err := Evaluate(group, closesToBars(98, 99, 101))
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = Evaluate(group, closesToBars(101, 102, 103))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestEvaluateIndicatorVsIndicator(t *testing.T) {
	closes := make([]float64, 60)
	for i := range closes {
		closes[i] = 100 + float64(i)
	}
	slow := ref("sma", map[string]float64{"period": 50})
	group := RuleGroup{
		Combinator: "AND",
		Rules: []Rule{{
			LHS: ref("sma", map[string]float64{"period": 10}),
			Op:  ">",
			RHS: RuleRHS{Kind: "indicator", Indicator: &slow},
		}},
	}
	ok, err := Evaluate(group, closesToBars(closes...))
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestEvaluateCombinators(t *testing.T) {
	trueRule := Rule{LHS: ref("price", nil), Op: ">", RHS: RuleRHS{Kind: "value", Value: 0}}
	falseRule := Rule{LHS: ref("price", nil), Op: "<", RHS: RuleRHS{Kind: "value", Value: 0}}

	bars := closesToBars(10, 11)

	ok, _ := Evaluate(RuleGroup{Combinator: "AND", Rules: []Rule{trueRule, falseRule}}, bars)
	assert.False(t, ok)

	ok, _ = Evaluate(RuleGroup{Combinator: "OR", Rules: []Rule{trueRule, falseRule}}, bars)
	assert.True(t, ok)
}

func TestEvaluateEmptyGroupNeverFires(t *testing.T) {
	ok, err := Evaluate(RuleGroup{Combinator: "AND"}, closesToBars(1, 2))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestEvaluateWarmupNaNIsFalse(t *testing.T) {
	group := RuleGroup{
		Combinator: "AND",
		Rules: []Rule{{
			LHS: ref("sma", map[string]float64{"period": 50}),
			Op:  ">",
			RHS: RuleRHS{Kind: "value", Value: 1},
		}},
	}
	ok, err := Evaluate(group, closesToBars(1, 2, 3, 4, 5))
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestParseConfigFromBuilderJSON(t *testing.T) {
	raw := []byte(`{
		"id": "x", "name": "RSI Dip", "description": "", "risk": "medium", "tags": [],
		"entry": {"combinator": "AND", "rules": [
			{"id": "r1", "lhs": {"kind": "rsi", "params": {"period": 14}}, "op": "<", "rhs": {"kind": "value", "value": 30}}
		]},
		"exit": {"combinator": "OR", "rules": [
			{"id": "r2", "lhs": {"kind": "rsi", "params": {"period": 14}}, "op": ">", "rhs": {"kind": "value", "value": 70}}
		]},
		"stopLossPct": 2, "takeProfitPct": 4, "createdAt": "2026-06-12"
	}`)

	cfg, err := ParseConfig(raw)
	require.NoError(t, err)
	assert.Equal(t, "RSI Dip", cfg.Name)
	assert.Len(t, cfg.Entry.Rules, 1)
	assert.Equal(t, "rsi", cfg.Entry.Rules[0].LHS.Kind)
	assert.Equal(t, 2.0, cfg.StopLossPct)
	assert.Equal(t, 4.0, cfg.TakeProfitPct)
	assert.Greater(t, MaxLookback(cfg), 14)
}

func TestParseConfigRejectsEmptyEntry(t *testing.T) {
	_, err := ParseConfig([]byte(`{"name":"x","entry":{"combinator":"AND","rules":[]}}`))
	assert.Error(t, err)
}
