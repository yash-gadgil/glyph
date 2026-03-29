package handlers

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	se "github.com/yash-gadgil/glyph/services/user/strategyengine"
	"go.uber.org/zap"
)

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

func TestGenerateStrategyFallbackIsDeployable(t *testing.T) {
	h := NewAdvisorHandler(nil, nil, nil, zap.NewNop())

	res, err := h.GenerateStrategy(context.Background(), &advisorpb.GenerateStrategyRequest{UserId: "user-1"})
	require.NoError(t, err)
	assert.NotEmpty(t, res.Name)
	assert.NotEmpty(t, res.Rationale)

	_, err = se.ParseConfig([]byte(res.ConfigJson))
	require.NoError(t, err)
}

func TestParseSelectionExtractsJSON(t *testing.T) {
	templates := strategyTemplates()
	out := "Sure! Here is my pick:\n{\"template\": \"sma_crossover\", \"rationale\": \"Reduces single-name timing risk.\"}\nThanks."

	key, rationale, ok := parseSelection(out, templates)
	require.True(t, ok)
	assert.Equal(t, "sma_crossover", key)
	assert.Equal(t, "Reduces single-name timing risk.", rationale)
}
