package strategyengine

import (
	"encoding/json"
	"fmt"
)

type IndicatorRef struct {
	Kind   string             `json:"kind"`
	Params map[string]float64 `json:"params"`
}

type RuleRHS struct {
	Kind      string        `json:"kind"`
	Value     float64       `json:"value"`
	Indicator *IndicatorRef `json:"indicator"`
}

type Rule struct {
	LHS IndicatorRef `json:"lhs"`
	Op  string       `json:"op"`
	RHS RuleRHS      `json:"rhs"`
}

type RuleGroup struct {
	Combinator string `json:"combinator"`
	Rules      []Rule `json:"rules"`
}

type Config struct {
	Name          string    `json:"name"`
	Entry         RuleGroup `json:"entry"`
	Exit          RuleGroup `json:"exit"`
	StopLossPct   float64   `json:"stopLossPct"`
	TakeProfitPct float64   `json:"takeProfitPct"`
}

func ParseConfig(raw []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return nil, fmt.Errorf("strategy config: %w", err)
	}
	if len(cfg.Entry.Rules) == 0 {
		return nil, fmt.Errorf("strategy config: entry rules are empty")
	}
	return &cfg, nil
}

func (r IndicatorRef) param(key string, fallback float64) float64 {
	if v, ok := r.Params[key]; ok && v > 0 {
		return v
	}
	return fallback
}
