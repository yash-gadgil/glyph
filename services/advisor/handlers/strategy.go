package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/yash-gadgil/glyph/pkg/logger"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type indicatorRef struct {
	Kind   string             `json:"kind"`
	Params map[string]float64 `json:"params"`
}

type ruleRHS struct {
	Kind      string        `json:"kind"`
	Value     float64       `json:"value,omitempty"`
	Indicator *indicatorRef `json:"indicator,omitempty"`
}

type rule struct {
	LHS indicatorRef `json:"lhs"`
	Op  string       `json:"op"`
	RHS ruleRHS      `json:"rhs"`
}

type ruleGroup struct {
	Combinator string `json:"combinator"`
	Rules      []rule `json:"rules"`
}

type genStrategy struct {
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	Risk          string    `json:"risk"`
	Tags          []string  `json:"tags"`
	Entry         ruleGroup `json:"entry"`
	Exit          ruleGroup `json:"exit"`
	StopLossPct   float64   `json:"stopLossPct"`
	TakeProfitPct float64   `json:"takeProfitPct"`
}

type strategyTemplate struct {
	Key   string
	Name  string
	Blurb string
	build func() genStrategy
}

func ind(kind string, params map[string]float64) indicatorRef {
	return indicatorRef{Kind: kind, Params: params}
}

func ruleVal(lhs indicatorRef, op string, v float64) rule {
	return rule{LHS: lhs, Op: op, RHS: ruleRHS{Kind: "value", Value: v}}
}

func ruleInd(lhs indicatorRef, op string, rhs indicatorRef) rule {
	r := rhs
	return rule{LHS: lhs, Op: op, RHS: ruleRHS{Kind: "indicator", Indicator: &r}}
}

func grp(comb string, rules ...rule) ruleGroup {
	return ruleGroup{Combinator: comb, Rules: rules}
}

func strategyTemplates() []strategyTemplate {
	return []strategyTemplate{
		{
			Key:   "rsi_dip_buyer",
			Name:  "RSI Dip Buyer",
			Blurb: "Mean reversion. Buys when RSI(14) dips below 40 (pulling back) and exits when it recovers above 60. Suits liquid names you want to accumulate on pullbacks.",
			build: func() genStrategy {
				return genStrategy{
					Risk:          "medium",
					Tags:          []string{"Mean Reversion", "RSI"},
					Entry:         grp("AND", ruleVal(ind("rsi", map[string]float64{"period": 14}), "<", 40)),
					Exit:          grp("OR", ruleVal(ind("rsi", map[string]float64{"period": 14}), ">", 60)),
					StopLossPct:   2,
					TakeProfitPct: 3,
				}
			},
		},
		{
			Key:   "momentum_breakout",
			Name:  "Momentum Breakout",
			Blurb: "Momentum. Enters when price crosses above the upper Bollinger band (20, 2 sigma) and exits when it falls back through the middle band. Suits trending, volatile names.",
			build: func() genStrategy {
				return genStrategy{
					Risk:          "high",
					Tags:          []string{"Momentum", "Breakout", "Bollinger"},
					Entry:         grp("AND", ruleInd(ind("price", map[string]float64{}), "crosses_above", ind("bbands_upper", map[string]float64{"period": 20, "stddev": 2}))),
					Exit:          grp("OR", ruleInd(ind("price", map[string]float64{}), "crosses_below", ind("bbands_middle", map[string]float64{"period": 20}))),
					StopLossPct:   2,
					TakeProfitPct: 5,
				}
			},
		},
		{
			Key:   "sma_crossover",
			Name:  "SMA Crossover",
			Blurb: "Trend following. Enters when the fast SMA(20) crosses above the slow SMA(50) and exits on the cross back below. Suits riding established trends and reducing single-name timing risk.",
			build: func() genStrategy {
				return genStrategy{
					Risk:          "medium",
					Tags:          []string{"Trend", "Moving Average"},
					Entry:         grp("AND", ruleInd(ind("sma", map[string]float64{"period": 20}), "crosses_above", ind("sma", map[string]float64{"period": 50}))),
					Exit:          grp("OR", ruleInd(ind("sma", map[string]float64{"period": 20}), "crosses_below", ind("sma", map[string]float64{"period": 50}))),
					StopLossPct:   3,
					TakeProfitPct: 6,
				}
			},
		},
	}
}

func (h *AdvisorHandler) GenerateStrategy(ctx context.Context, req *advisorpb.GenerateStrategyRequest) (*advisorpb.StrategySuggestion, error) {
	log := logger.WithContextFields(ctx, h.log).With(logger.Action("generate_strategy"))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	templates := strategyTemplates()

	snapshot, _, err := buildSnapshot(ctx, h.portfolio, req.UserId)
	if err != nil {
		log.Warn("snapshot_unavailable_using_generic_selection", zap.Error(err))
		snapshot = "No portfolio snapshot available."
	}

	key, rationale := h.selectTemplate(ctx, snapshot, templates)
	tmpl := templates[0]
	for _, t := range templates {
		if t.Key == key {
			tmpl = t
			break
		}
	}
	if rationale == "" {
		rationale = tmpl.Blurb
	}

	gs := tmpl.build()
	gs.Name = fmt.Sprintf("%s %s", tmpl.Name, shortID())
	gs.Description = rationale

	if len(gs.Entry.Rules) == 0 {
		log.Error("generated_strategy_invalid", zap.String("template", tmpl.Key))
		return nil, status.Error(codes.Internal, "generated strategy failed validation")
	}

	configJSON, err := json.Marshal(gs)
	if err != nil {
		log.Error("marshal_strategy_failed", zap.Error(err))
		return nil, status.Error(codes.Internal, "could not build strategy")
	}

	return &advisorpb.StrategySuggestion{
		Name:       gs.Name,
		ConfigJson: string(configJSON),
		Rationale:  rationale,
		Template:   tmpl.Key,
	}, nil
}

func (h *AdvisorHandler) selectTemplate(ctx context.Context, snapshot string, templates []strategyTemplate) (string, string) {
	fallback := templates[0].Key
	if h.llm == nil {
		return fallback, ""
	}

	var menu strings.Builder
	for _, t := range templates {
		fmt.Fprintf(&menu, "- %s: %s\n", t.Key, t.Blurb)
	}

	system := "You pick one trading strategy template for a user's portfolio and explain why. " +
		"Reply with ONLY a JSON object and nothing else, in the form " +
		`{"template": "<one of the listed keys>", "rationale": "<one or two sentences referencing the portfolio>"}.`
	prompt := snapshot + "\nAvailable templates:\n" + menu.String() +
		"\nChoose the single best template key for this portfolio and give a short rationale. Respond with only the JSON object."

	out, err := h.llm.CompleteShort(ctx, system, prompt, 160)
	if err != nil {
		return fallback, ""
	}

	key, rationale, ok := parseSelection(out, templates)
	if !ok {
		return fallback, ""
	}
	return key, rationale
}

func parseSelection(out string, templates []strategyTemplate) (string, string, bool) {
	start := strings.Index(out, "{")
	end := strings.LastIndex(out, "}")
	if start < 0 || end <= start {
		return "", "", false
	}

	var sel struct {
		Template  string `json:"template"`
		Rationale string `json:"rationale"`
	}
	if err := json.Unmarshal([]byte(out[start:end+1]), &sel); err != nil {
		return "", "", false
	}

	for _, t := range templates {
		if t.Key == sel.Template {
			return sel.Template, strings.TrimSpace(sel.Rationale), true
		}
	}
	return "", "", false
}

func shortID() string {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "x"
	}
	return hex.EncodeToString(b)
}
