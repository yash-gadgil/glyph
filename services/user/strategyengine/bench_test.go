package strategyengine

import (
	"math"
	"testing"
	"time"
)

func benchBars(n int) []Bar {
	bars := make([]Bar, n)
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	price := 100.0
	for i := 0; i < n; i++ {
		price += math.Sin(float64(i)/7)*0.5 + math.Cos(float64(i)/13)*0.3
		if price < 1 {
			price = 1
		}
		bars[i] = Bar{
			Time:   base.Add(time.Duration(i) * time.Minute),
			Open:   price,
			High:   price + 1,
			Low:    price - 1,
			Close:  price,
			Volume: 1000,
		}
	}
	return bars
}

func rsiSmaCrossConfig() *Config {
	sma20 := IndicatorRef{Kind: "sma", Params: map[string]float64{"period": 20}}
	return &Config{
		Name: "rsi+smacross",
		Entry: RuleGroup{Combinator: "AND", Rules: []Rule{
			{LHS: IndicatorRef{Kind: "rsi", Params: map[string]float64{"period": 14}}, Op: "<", RHS: RuleRHS{Kind: "value", Value: 35}},
			{LHS: IndicatorRef{Kind: "sma", Params: map[string]float64{"period": 5}}, Op: "crosses_above", RHS: RuleRHS{Kind: "indicator", Indicator: &sma20}},
		}},
		Exit: RuleGroup{Combinator: "OR", Rules: []Rule{
			{LHS: IndicatorRef{Kind: "rsi", Params: map[string]float64{"period": 14}}, Op: ">", RHS: RuleRHS{Kind: "value", Value: 70}},
		}},
		StopLossPct:   2,
		TakeProfitPct: 4,
	}
}

func BenchmarkEvaluate(b *testing.B) {
	cfg := rsiSmaCrossConfig()
	bars := benchBars(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Evaluate(cfg.Entry, bars)
	}
}

func BenchmarkSeriesMACD(b *testing.B) {
	ref := IndicatorRef{Kind: "macd_line", Params: map[string]float64{"fast": 12, "slow": 26, "signal": 9}}
	bars := benchBars(1000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = Series(ref, bars)
	}
}

func BenchmarkRunBacktest_Daily126(b *testing.B) {
	cfg := rsiSmaCrossConfig()
	bars := benchBars(126)
	params := BacktestParams{InitialCapitalCents: 10_000_000, PositionSizeCents: 1_000_000, Timeframe: "DAY"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := RunBacktest(cfg, bars, params); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRunBacktest_Min5000(b *testing.B) {
	cfg := rsiSmaCrossConfig()
	bars := benchBars(5000)
	params := BacktestParams{InitialCapitalCents: 10_000_000, PositionSizeCents: 1_000_000, Timeframe: "MIN"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := RunBacktest(cfg, bars, params); err != nil {
			b.Fatal(err)
		}
	}
}
