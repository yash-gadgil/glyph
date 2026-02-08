package strategyengine

import (
	"fmt"
	"math"
	"time"
)

type Bar struct {
	Time   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	VWAP   float64
}

func Series(ref IndicatorRef, bars []Bar) ([]float64, error) {
	closes := make([]float64, len(bars))
	for i, b := range bars {
		closes[i] = b.Close
	}

	switch ref.Kind {
	case "price":
		return closes, nil

	case "sma":
		return sma(closes, int(ref.param("period", 20))), nil

	case "ema":
		return ema(closes, int(ref.param("period", 20))), nil

	case "rsi":
		return rsi(closes, int(ref.param("period", 14))), nil

	case "macd_line":
		return macdLine(closes, int(ref.param("fast", 12)), int(ref.param("slow", 26))), nil

	case "macd_signal":
		line := macdLine(closes, int(ref.param("fast", 12)), int(ref.param("slow", 26)))
		return emaOverNaN(line, int(ref.param("signal", 9))), nil

	case "macd_histogram":
		line := macdLine(closes, int(ref.param("fast", 12)), int(ref.param("slow", 26)))
		signal := emaOverNaN(line, int(ref.param("signal", 9)))
		return sub(line, signal), nil

	case "bbands_upper", "bbands_middle", "bbands_lower":
		period := int(ref.param("period", 20))
		dev := ref.param("stddev", 2)
		mid := sma(closes, period)
		if ref.Kind == "bbands_middle" {
			return mid, nil
		}
		sd := rollingStdev(closes, period)
		sign := 1.0
		if ref.Kind == "bbands_lower" {
			sign = -1.0
		}
		out := make([]float64, len(closes))
		for i := range out {
			out[i] = mid[i] + sign*dev*sd[i]
		}
		return out, nil

	case "atr":
		return atr(bars, int(ref.param("period", 14))), nil

	case "volume":
		out := make([]float64, len(bars))
		for i, b := range bars {
			out[i] = b.Volume
		}
		return out, nil

	case "vwap":
		out := make([]float64, len(bars))
		for i, b := range bars {
			if b.VWAP > 0 {
				out[i] = b.VWAP
			} else {
				out[i] = (b.High + b.Low + b.Close) / 3
			}
		}
		return out, nil

	case "stoch_k":
		return stochK(bars, int(ref.param("period", 14))), nil

	case "stoch_d":
		return sma(stochK(bars, int(ref.param("period", 14))), 3), nil
	}

	return nil, fmt.Errorf("unknown indicator %q", ref.Kind)
}

func nanSlice(n int) []float64 {
	out := make([]float64, n)
	for i := range out {
		out[i] = math.NaN()
	}
	return out
}

func sma(values []float64, period int) []float64 {
	out := nanSlice(len(values))
	if period < 1 || len(values) < period {
		return out
	}
	var sum float64
	for i, v := range values {
		if math.IsNaN(v) {
			return smaSkipNaN(values, period)
		}
		sum += v
		if i >= period {
			sum -= values[i-period]
		}
		if i >= period-1 {
			out[i] = sum / float64(period)
		}
	}
	return out
}

func smaSkipNaN(values []float64, period int) []float64 {
	out := nanSlice(len(values))
	start := 0
	for start < len(values) && math.IsNaN(values[start]) {
		start++
	}
	var sum float64
	for i := start; i < len(values); i++ {
		sum += values[i]
		if i-start >= period {
			sum -= values[i-period]
		}
		if i-start >= period-1 {
			out[i] = sum / float64(period)
		}
	}
	return out
}

func ema(values []float64, period int) []float64 {
	out := nanSlice(len(values))
	if period < 1 || len(values) < period {
		return out
	}
	var sum float64
	for i := 0; i < period; i++ {
		sum += values[i]
	}
	prev := sum / float64(period)
	out[period-1] = prev
	k := 2.0 / (float64(period) + 1.0)
	for i := period; i < len(values); i++ {
		prev = (values[i]-prev)*k + prev
		out[i] = prev
	}
	return out
}

func emaOverNaN(values []float64, period int) []float64 {
	out := nanSlice(len(values))
	start := 0
	for start < len(values) && math.IsNaN(values[start]) {
		start++
	}
	if period < 1 || len(values)-start < period {
		return out
	}
	tail := ema(values[start:], period)
	copy(out[start:], tail)
	return out
}

func rsi(closes []float64, period int) []float64 {
	out := nanSlice(len(closes))
	if period < 1 || len(closes) < period+1 {
		return out
	}
	var gain, loss float64
	for i := 1; i <= period; i++ {
		diff := closes[i] - closes[i-1]
		if diff >= 0 {
			gain += diff
		} else {
			loss -= diff
		}
	}
	avgGain := gain / float64(period)
	avgLoss := loss / float64(period)
	out[period] = rsiValue(avgGain, avgLoss)
	for i := period + 1; i < len(closes); i++ {
		diff := closes[i] - closes[i-1]
		g, l := 0.0, 0.0
		if diff >= 0 {
			g = diff
		} else {
			l = -diff
		}
		avgGain = (avgGain*float64(period-1) + g) / float64(period)
		avgLoss = (avgLoss*float64(period-1) + l) / float64(period)
		out[i] = rsiValue(avgGain, avgLoss)
	}
	return out
}

func rsiValue(avgGain, avgLoss float64) float64 {
	if avgLoss == 0 {
		return 100
	}
	return 100 - 100/(1+avgGain/avgLoss)
}

func macdLine(closes []float64, fast, slow int) []float64 {
	return sub(ema(closes, fast), ema(closes, slow))
}

func sub(a, b []float64) []float64 {
	out := nanSlice(len(a))
	for i := range a {
		out[i] = a[i] - b[i]
	}
	return out
}

func rollingStdev(values []float64, period int) []float64 {
	out := nanSlice(len(values))
	if period < 2 || len(values) < period {
		return out
	}
	for i := period - 1; i < len(values); i++ {
		window := values[i-period+1 : i+1]
		var mean float64
		for _, v := range window {
			mean += v
		}
		mean /= float64(period)
		var variance float64
		for _, v := range window {
			variance += (v - mean) * (v - mean)
		}
		out[i] = math.Sqrt(variance / float64(period))
	}
	return out
}

func atr(bars []Bar, period int) []float64 {
	out := nanSlice(len(bars))
	if period < 1 || len(bars) < period+1 {
		return out
	}
	tr := make([]float64, len(bars))
	tr[0] = bars[0].High - bars[0].Low
	for i := 1; i < len(bars); i++ {
		hl := bars[i].High - bars[i].Low
		hc := math.Abs(bars[i].High - bars[i-1].Close)
		lc := math.Abs(bars[i].Low - bars[i-1].Close)
		tr[i] = math.Max(hl, math.Max(hc, lc))
	}
	var sum float64
	for i := 1; i <= period; i++ {
		sum += tr[i]
	}
	prev := sum / float64(period)
	out[period] = prev
	for i := period + 1; i < len(bars); i++ {
		prev = (prev*float64(period-1) + tr[i]) / float64(period)
		out[i] = prev
	}
	return out
}

func stochK(bars []Bar, period int) []float64 {
	out := nanSlice(len(bars))
	if period < 1 || len(bars) < period {
		return out
	}
	for i := period - 1; i < len(bars); i++ {
		hi, lo := math.Inf(-1), math.Inf(1)
		for j := i - period + 1; j <= i; j++ {
			hi = math.Max(hi, bars[j].High)
			lo = math.Min(lo, bars[j].Low)
		}
		if hi == lo {
			out[i] = 50
			continue
		}
		out[i] = (bars[i].Close - lo) / (hi - lo) * 100
	}
	return out
}
