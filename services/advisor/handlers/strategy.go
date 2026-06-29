package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/advisor/cache"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	jobRunning   = "running"
	jobSucceeded = "succeeded"
	jobFailed    = "failed"

	maxGenAttempts    = 2
	generationTimeout = 90 * time.Second
	backtestSymbol    = "AAPL"
)

type indicatorRef struct {
	Kind   string             `json:"kind"`
	Params map[string]float64 `json:"params"`
}

type ruleRHS struct {
	Kind      string             `json:"kind"`
	Value     float64            `json:"value,omitempty"`
	Indicator *indicatorRef      `json:"indicator,omitempty"`
	Params    map[string]float64 `json:"params,omitempty"`
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

var validIndicators = map[string]bool{
	"price": true, "sma": true, "ema": true, "rsi": true,
	"macd_line": true, "macd_signal": true, "macd_histogram": true,
	"bbands_upper": true, "bbands_middle": true, "bbands_lower": true,
	"atr": true, "volume": true, "vwap": true, "stoch_k": true, "stoch_d": true,
}

var validOps = map[string]bool{
	">": true, "<": true, ">=": true, "<=": true,
	"crosses_above": true, "crosses_below": true,
}

var validCombinators = map[string]bool{"AND": true, "OR": true}

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

func (h *AdvisorHandler) StartStrategyGeneration(ctx context.Context, req *advisorpb.StartStrategyGenerationRequest) (*advisorpb.StrategyJob, error) {
	log := logger.WithContextFields(ctx, h.log).With(logger.Action("start_strategy_generation"))

	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	if h.cache != nil && h.cache.Enabled() {
		if job, err := h.cache.GetJob(ctx, req.UserId); err == nil && job != nil && job.State == jobRunning {
			return jobToProto(job), nil
		}

		ok, err := h.cache.AcquireJobLock(ctx, req.UserId, generationTimeout)
		if err != nil {
			log.Warn("job_lock_error", zap.Error(err))
		}
		if !ok {
			if job, _ := h.cache.GetJob(ctx, req.UserId); job != nil {
				return jobToProto(job), nil
			}
			return &advisorpb.StrategyJob{State: jobRunning}, nil
		}

		job := &cache.StratJob{State: jobRunning, StartedAt: time.Now().UTC()}
		if err := h.cache.SetJob(ctx, req.UserId, job); err != nil {
			log.Warn("job_persist_error", zap.Error(err))
		}

		genCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), generationTimeout)
		go func() {
			defer cancel()
			defer h.cache.ReleaseJobLock(context.Background(), req.UserId)
			result := h.generateOnce(genCtx, req.UserId, log)
			if err := h.cache.SetJob(context.Background(), req.UserId, result); err != nil {
				log.Error("job_result_persist_error", zap.Error(err))
			}
		}()

		return jobToProto(job), nil
	}

	return jobToProto(h.generateOnce(ctx, req.UserId, log)), nil
}

func (h *AdvisorHandler) GetStrategyJob(ctx context.Context, req *advisorpb.GetStrategyJobRequest) (*advisorpb.StrategyJob, error) {
	if req.UserId == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if h.cache == nil {
		return &advisorpb.StrategyJob{}, nil
	}
	job, err := h.cache.GetJob(ctx, req.UserId)
	if err != nil {
		return nil, status.Error(codes.Internal, "could not read strategy job")
	}
	if job == nil {
		return &advisorpb.StrategyJob{}, nil
	}
	return jobToProto(job), nil
}

func (h *AdvisorHandler) generateOnce(ctx context.Context, userID string, log *zap.Logger) *cache.StratJob {
	started := time.Now().UTC()

	snapshot, _, err := buildSnapshot(ctx, h.portfolio, userID)
	if err != nil {
		log.Warn("snapshot_unavailable_using_generic_context", zap.Error(err))
		snapshot = "No portfolio snapshot available."
	}

	gs, summary, err := h.authorStrategy(ctx, userID, snapshot, log)
	if err != nil {
		log.Warn("strategy_generation_failed", zap.Error(err))
		return &cache.StratJob{State: jobFailed, Error: err.Error(), StartedAt: started, UpdatedAt: time.Now().UTC()}
	}

	configJSON, err := json.Marshal(gs)
	if err != nil {
		return &cache.StratJob{State: jobFailed, Error: "could not encode strategy", StartedAt: started, UpdatedAt: time.Now().UTC()}
	}

	log.Info("strategy_generation_succeeded", logger.KV("name", gs.Name))
	return &cache.StratJob{
		State:      jobSucceeded,
		Name:       gs.Name,
		ConfigJSON: string(configJSON),
		Rationale:  gs.Description,
		Backtest:   summary,
		StartedAt:  started,
		UpdatedAt:  time.Now().UTC(),
	}
}

func (h *AdvisorHandler) authorStrategy(ctx context.Context, userID, snapshot string, log *zap.Logger) (genStrategy, *cache.BacktestSummary, error) {
	if h.llm == nil {
		return h.fallbackStrategy(), nil, nil
	}

	var lastErr error
	feedback := ""
	for attempt := 0; attempt < maxGenAttempts; attempt++ {
		out, err := h.llm.CompleteShort(ctx, authorSystem, buildAuthorPrompt(snapshot, feedback), 1024)
		if err != nil {
			return genStrategy{}, nil, fmt.Errorf("model error: %w", err)
		}

		gs, perr := parseGenStrategy(out)
		if perr != nil {
			lastErr, feedback = perr, perr.Error()
			log.Warn("strategy_parse_rejected", zap.Int("attempt", attempt), zap.Error(perr))
			continue
		}
		normalizeStrategy(&gs)
		if verr := validate(gs); verr != nil {
			lastErr, feedback = verr, verr.Error()
			log.Warn("strategy_validation_rejected", zap.Int("attempt", attempt), zap.Error(verr))
			continue
		}

		gs.Name = fmt.Sprintf("%s %s", strings.TrimSpace(gs.Name), shortID())

		summary, retryReason, infraErr := h.backtestGuard(ctx, userID, gs)
		if retryReason != "" {
			lastErr, feedback = errors.New(retryReason), retryReason
			log.Warn("strategy_backtest_rejected", zap.Int("attempt", attempt), zap.String("reason", retryReason))
			continue
		}
		if infraErr != nil {
			log.Warn("strategy_backtest_unavailable", zap.Error(infraErr))
		}
		return gs, summary, nil
	}

	return genStrategy{}, nil, fmt.Errorf("strategy failed validation after %d attempts: %w", maxGenAttempts, lastErr)
}

func (h *AdvisorHandler) backtestGuard(ctx context.Context, userID string, gs genStrategy) (*cache.BacktestSummary, string, error) {
	if h.strategy == nil {
		return nil, "", nil
	}

	configJSON, err := json.Marshal(gs)
	if err != nil {
		return nil, "", err
	}

	end := time.Now().UTC()
	start := end.AddDate(-1, 0, 0)
	resp, err := h.strategy.RunBacktest(ctx, &strategypb.BacktestRequest{
		UserId:              userID,
		ConfigJson:          string(configJSON),
		Symbol:              backtestSymbol,
		Timeframe:           "DAY",
		Start:               start.Format("2006-01-02"),
		End:                 end.Format("2006-01-02"),
		InitialCapitalCents: 10_000_000,
		PositionSizeCents:   1_000_000,
	})
	if err != nil {
		return nil, "", err
	}
	if resp.NumTrades == 0 {
		return nil, "the strategy produced zero trades over a one year backtest; loosen the entry conditions so it actually triggers", nil
	}

	return &cache.BacktestSummary{
		TotalReturnPct: resp.TotalReturnPct,
		MaxDrawdownPct: resp.MaxDrawdownPct,
		Sharpe:         resp.Sharpe,
		WinRate:        resp.WinRate,
		ProfitFactor:   resp.ProfitFactor,
		NumTrades:      resp.NumTrades,
	}, "", nil
}

func (h *AdvisorHandler) fallbackStrategy() genStrategy {
	tmpl := strategyTemplates()[0]
	gs := tmpl.build()
	gs.Name = fmt.Sprintf("%s %s", tmpl.Name, shortID())
	gs.Description = tmpl.Blurb
	return gs
}

func parseGenStrategy(out string) (genStrategy, error) {
	start := strings.Index(out, "{")
	end := strings.LastIndex(out, "}")
	if start < 0 || end <= start {
		return genStrategy{}, errors.New("model did not return a JSON object")
	}
	var gs genStrategy
	if err := json.Unmarshal([]byte(out[start:end+1]), &gs); err != nil {
		return genStrategy{}, fmt.Errorf("model output was not valid strategy JSON: %v", err)
	}
	return gs, nil
}

func normalizeStrategy(gs *genStrategy) {
	for i := range gs.Entry.Rules {
		normalizeRHS(&gs.Entry.Rules[i].RHS)
	}
	for i := range gs.Exit.Rules {
		normalizeRHS(&gs.Exit.Rules[i].RHS)
	}
}

func normalizeRHS(r *ruleRHS) {
	if r.Indicator != nil {
		r.Kind = "indicator"
		r.Params = nil
		return
	}
	switch r.Kind {
	case "", "value":
		r.Kind = "value"
	case "indicator":
	default:
		if validIndicators[r.Kind] {
			r.Indicator = &indicatorRef{Kind: r.Kind, Params: r.Params}
			r.Kind = "indicator"
			r.Params = nil
		}
	}
}

func validate(gs genStrategy) error {
	if strings.TrimSpace(gs.Name) == "" {
		return errors.New("name is required")
	}
	if err := validateGroup("entry", gs.Entry, true); err != nil {
		return err
	}
	if err := validateGroup("exit", gs.Exit, false); err != nil {
		return err
	}
	if gs.StopLossPct < 0 || gs.StopLossPct > 50 {
		return fmt.Errorf("stopLossPct must be between 0 and 50, got %v", gs.StopLossPct)
	}
	if gs.TakeProfitPct < 0 || gs.TakeProfitPct > 100 {
		return fmt.Errorf("takeProfitPct must be between 0 and 100, got %v", gs.TakeProfitPct)
	}
	if len(gs.Exit.Rules) == 0 && gs.StopLossPct == 0 && gs.TakeProfitPct == 0 {
		return errors.New("strategy needs an exit: provide exit rules or a non-zero stopLossPct/takeProfitPct")
	}
	return nil
}

func validateGroup(label string, g ruleGroup, requireRules bool) error {
	if len(g.Rules) == 0 {
		if requireRules {
			return fmt.Errorf("%s must contain at least one rule", label)
		}
		return nil
	}
	if !validCombinators[strings.ToUpper(g.Combinator)] {
		return fmt.Errorf("%s combinator %q must be AND or OR", label, g.Combinator)
	}
	for i, r := range g.Rules {
		if !validIndicators[r.LHS.Kind] {
			return fmt.Errorf("%s rule %d uses unknown indicator %q", label, i, r.LHS.Kind)
		}
		if !validOps[r.Op] {
			return fmt.Errorf("%s rule %d uses unknown operator %q", label, i, r.Op)
		}
		switch r.RHS.Kind {
		case "value", "":
		case "indicator":
			if r.RHS.Indicator == nil {
				return fmt.Errorf("%s rule %d declares an indicator rhs but provides none", label, i)
			}
			if !validIndicators[r.RHS.Indicator.Kind] {
				return fmt.Errorf("%s rule %d rhs uses unknown indicator %q", label, i, r.RHS.Indicator.Kind)
			}
		default:
			return fmt.Errorf("%s rule %d rhs kind %q must be value or indicator", label, i, r.RHS.Kind)
		}
	}
	return nil
}

func jobToProto(j *cache.StratJob) *advisorpb.StrategyJob {
	out := &advisorpb.StrategyJob{
		State:      j.State,
		Name:       j.Name,
		ConfigJson: j.ConfigJSON,
		Rationale:  j.Rationale,
		Error:      j.Error,
	}
	if !j.StartedAt.IsZero() {
		out.StartedAt = j.StartedAt.UTC().Format(time.RFC3339)
	}
	if !j.UpdatedAt.IsZero() {
		out.UpdatedAt = j.UpdatedAt.UTC().Format(time.RFC3339)
	}
	if j.Backtest != nil {
		out.Backtest = &advisorpb.BacktestSummary{
			TotalReturnPct: j.Backtest.TotalReturnPct,
			MaxDrawdownPct: j.Backtest.MaxDrawdownPct,
			Sharpe:         j.Backtest.Sharpe,
			WinRate:        j.Backtest.WinRate,
			ProfitFactor:   j.Backtest.ProfitFactor,
			NumTrades:      j.Backtest.NumTrades,
		}
	}
	return out
}

const authorSystem = `You are a quantitative strategy author. You design one rule-based trading strategy and return it as JSON only, no markdown.

Return ONLY a JSON object in exactly this shape:
{
  "name": "<short strategy name>",
  "description": "<one or two sentences on the idea and why it suits the portfolio>",
  "risk": "low | medium | high",
  "tags": ["<tag>"],
  "entry": { "combinator": "AND | OR", "rules": [ <rule> ] },
  "exit": { "combinator": "AND | OR", "rules": [ <rule> ] },
  "stopLossPct": <number 0-50>,
  "takeProfitPct": <number 0-100>
}

A <rule> is { "lhs": <indicator>, "op": "<operator>", "rhs": <rhs> }.
An <indicator> is { "kind": "<indicator kind>", "params": { "<param>": <number> } }.
A <rhs> is either { "kind": "value", "value": <number> } or { "kind": "indicator", "indicator": <indicator> }.

Allowed indicator kinds: price, sma, ema, rsi, macd_line, macd_signal, macd_histogram, bbands_upper, bbands_middle, bbands_lower, atr, volume, vwap, stoch_k, stoch_d.
Params: sma/ema/rsi/atr/stoch_k/stoch_d use {"period": N}; macd_* use {"fast":12,"slow":26,"signal":9}; bbands_* use {"period":20,"stddev":2}; price/volume/vwap take no params.
Allowed operators: >, <, >=, <=, crosses_above, crosses_below.

To compare against a number use {"kind":"value","value":N}. To compare against another indicator you MUST wrap it: {"kind":"indicator","indicator":{"kind":"sma","params":{"period":50}}}. Never put an indicator kind directly in the rhs "kind" field.

Complete example:
{"name":"SMA Crossover","description":"Rides established uptrends.","risk":"medium","tags":["Trend"],"entry":{"combinator":"AND","rules":[{"lhs":{"kind":"sma","params":{"period":20}},"op":"crosses_above","rhs":{"kind":"indicator","indicator":{"kind":"sma","params":{"period":50}}}}]},"exit":{"combinator":"OR","rules":[{"lhs":{"kind":"rsi","params":{"period":14}},"op":">","rhs":{"kind":"value","value":70}}]},"stopLossPct":3,"takeProfitPct":6}

Constraints: entry has 1 to 3 rules; always provide an exit via exit rules or a non-zero stopLossPct/takeProfitPct; keep it simple and tradeable.`

func buildAuthorPrompt(snapshot, feedback string) string {
	var b strings.Builder
	b.WriteString(snapshot)
	b.WriteString("\n\nAuthor one strategy tailored to this portfolio. Respond with only the JSON object.")
	if feedback != "" {
		b.WriteString("\n\nYour previous attempt was rejected for this reason, fix it: ")
		b.WriteString(feedback)
	}
	return b.String()
}

func shortID() string {
	b := make([]byte, 2)
	if _, err := rand.Read(b); err != nil {
		return "x"
	}
	return hex.EncodeToString(b)
}
