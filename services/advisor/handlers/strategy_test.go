package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	se "github.com/yash-gadgil/glyph/services/strategy/engine"
	"go.uber.org/zap"
)

type scriptedProvider struct {
	replies []string
	calls   int
}

func (p *scriptedProvider) Stream(ctx context.Context, system, prompt string, emit func(string) error) error {
	return nil
}

func (p *scriptedProvider) CompleteShort(ctx context.Context, system, prompt string, maxTokens int32) (string, error) {
	reply := p.replies[p.calls%len(p.replies)]
	p.calls++
	return reply, nil
}

func TestStrategyTemplatesParse(t *testing.T) {
	for _, tmpl := range strategyTemplates() {
		gs := tmpl.build()
		gs.Name = tmpl.Name
		raw, err := json.Marshal(gs)
		require.NoError(t, err, tmpl.Key)

		_, err = se.ParseConfig(raw)
		require.NoError(t, err, "template %s should parse", tmpl.Key)
	}
}

func TestValidateRejectsBadStrategies(t *testing.T) {
	valid := strategyTemplates()[0].build()
	valid.Name = "Valid"
	require.NoError(t, validate(valid))

	cases := map[string]func(genStrategy) genStrategy{
		"missing name": func(g genStrategy) genStrategy { g.Name = ""; return g },
		"empty entry":  func(g genStrategy) genStrategy { g.Entry = ruleGroup{Combinator: "AND"}; return g },
		"unknown indicator": func(g genStrategy) genStrategy {
			g.Entry = grp("AND", ruleVal(ind("supertrend", nil), "<", 30))
			return g
		},
		"unknown operator": func(g genStrategy) genStrategy {
			g.Entry = grp("AND", rule{LHS: ind("rsi", map[string]float64{"period": 14}), Op: "approaches", RHS: ruleRHS{Kind: "value", Value: 30}})
			return g
		},
		"insane stop loss": func(g genStrategy) genStrategy { g.StopLossPct = 999; return g },
		"no exit": func(g genStrategy) genStrategy {
			g.Exit = ruleGroup{}
			g.StopLossPct = 0
			g.TakeProfitPct = 0
			return g
		},
	}

	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			base := strategyTemplates()[0].build()
			base.Name = "Base"
			assert.Error(t, validate(mutate(base)), "expected %s to be rejected", name)
		})
	}
}

func TestNormalizeStrategyRewritesInlinedIndicatorRHS(t *testing.T) {
	raw := `{"name":"x","entry":{"combinator":"AND","rules":[{"lhs":{"kind":"sma","params":{"period":20}},"op":"crosses_above","rhs":{"kind":"sma","params":{"period":50}}}]},"stopLossPct":2}`

	gs, err := parseGenStrategy(raw)
	require.NoError(t, err)
	normalizeStrategy(&gs)

	require.NoError(t, validate(gs))
	rhs := gs.Entry.Rules[0].RHS
	assert.Equal(t, "indicator", rhs.Kind)
	require.NotNil(t, rhs.Indicator)
	assert.Equal(t, "sma", rhs.Indicator.Kind)
	assert.Equal(t, float64(50), rhs.Indicator.Params["period"])
	assert.Nil(t, rhs.Params)

	out, err := json.Marshal(gs)
	require.NoError(t, err)
	cfg, err := se.ParseConfig(out)
	require.NoError(t, err)
	assert.Equal(t, "indicator", cfg.Entry.Rules[0].RHS.Kind)
}

func TestAuthorStrategyRetriesOnInvalidThenSucceeds(t *testing.T) {
	broken := `here you go: {"name":"Bad","entry":{"combinator":"AND","rules":[{"lhs":{"kind":"unicorn"},"op":"<","rhs":{"kind":"value","value":30}}]}}`
	good := `{"name":"RSI Reversion","description":"Buys oversold dips.","risk":"medium","entry":{"combinator":"AND","rules":[{"lhs":{"kind":"rsi","params":{"period":14}},"op":"<","rhs":{"kind":"value","value":35}}]},"exit":{"combinator":"OR","rules":[{"lhs":{"kind":"rsi","params":{"period":14}},"op":">","rhs":{"kind":"value","value":65}}]},"stopLossPct":2,"takeProfitPct":4}`

	provider := &scriptedProvider{replies: []string{broken, good}}
	h := NewAdvisorHandler(nil, nil, provider, nil, zap.NewNop())

	gs, summary, err := h.authorStrategy(context.Background(), "user-1", "No portfolio snapshot.", zap.NewNop())
	require.NoError(t, err)
	assert.Equal(t, 2, provider.calls, "should retry once after the broken attempt")
	assert.Nil(t, summary)
	assert.Contains(t, gs.Name, "RSI Reversion")

	raw, err := json.Marshal(gs)
	require.NoError(t, err)
	_, err = se.ParseConfig(raw)
	require.NoError(t, err)
}

func TestAuthorStrategyFailsWhenAlwaysInvalid(t *testing.T) {
	provider := &scriptedProvider{replies: []string{"not json at all"}}
	h := NewAdvisorHandler(nil, nil, provider, nil, zap.NewNop())

	_, _, err := h.authorStrategy(context.Background(), "user-1", "snapshot", zap.NewNop())
	require.Error(t, err)
	assert.Equal(t, maxGenAttempts, provider.calls)
}

func TestStartStrategyGenerationFallbackIsDeployable(t *testing.T) {
	h := NewAdvisorHandler(nil, nil, nil, nil, zap.NewNop())

	job, err := h.StartStrategyGeneration(context.Background(), &advisorpb.StartStrategyGenerationRequest{UserId: "user-1"})
	require.NoError(t, err)
	assert.Equal(t, jobSucceeded, job.State)
	assert.NotEmpty(t, job.Name)

	_, err = se.ParseConfig([]byte(job.ConfigJson))
	require.NoError(t, err)
}
