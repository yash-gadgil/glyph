package handlers

import (
	"time"

	"github.com/yash-gadgil/glyph/services/advisor/cache"
)

const (
	regenThreshold = 0.15
	maxAge         = 24 * time.Hour
)

func changeScore(prev, cur cache.Fingerprint) (float64, bool) {
	base := prev.EquityCents
	if base <= 0 {
		base = cur.EquityCents
	}
	if base <= 0 {
		return 1, true
	}

	symbolsChanged := false
	var tradedCents int64
	seen := make(map[string]bool, len(cur.Positions))

	for sym, c := range cur.Positions {
		p, ok := prev.Positions[sym]
		if !ok {
			symbolsChanged = true
		}
		tradedCents += absI64(c.Qty-p.Qty) * c.AvgPriceCents
		seen[sym] = true
	}

	for sym, p := range prev.Positions {
		if !seen[sym] {
			symbolsChanged = true
			tradedCents += p.Qty * p.AvgPriceCents
		}
	}

	return float64(tradedCents) / float64(base), symbolsChanged
}

func shouldRegenerate(prev *cache.Entry, cur cache.Fingerprint) bool {
	if prev == nil {
		return true
	}
	if time.Since(prev.GeneratedAt) > maxAge {
		return true
	}
	score, symbolsChanged := changeScore(prev.Fingerprint, cur)
	return symbolsChanged || score >= regenThreshold
}

func absI64(v int64) int64 {
	if v < 0 {
		return -v
	}
	return v
}
