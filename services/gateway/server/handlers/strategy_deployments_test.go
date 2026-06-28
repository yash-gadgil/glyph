package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestDeployStrategyCreated(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("DeployStrategy", mock.Anything, mock.MatchedBy(func(req *strategypb.DeployStrategyRequest) bool {
		return req.StrategyId == "s1" && req.Symbol == "AAPL"
	})).Return(&strategypb.Deployment{Id: "d1", StrategyId: "s1", Symbol: "AAPL", Status: "RUNNING"}, nil)

	cfg := strategyConfig(strategy)
	body, _ := json.Marshal(map[string]any{"symbol": "AAPL", "position_size_cents": 1_000_000})
	req := withURLParam(handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/strategies/s1/deploy", bytes.NewReader(body)), "user-1"), "id", "s1")
	w := httptest.NewRecorder()

	cfg.DeployStrategy(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	strategy.AssertExpectations(t)
}

func TestGetDeploymentsReturnsList(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("GetDeployments", mock.Anything, &strategypb.UserSpecifier{UserId: "user-1"}).
		Return(&strategypb.DeploymentsResponse{
			Deployments: []*strategypb.Deployment{{Id: "d1", Symbol: "AAPL"}},
		}, nil)

	cfg := strategyConfig(strategy)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/strategies/deployments", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetDeployments(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Deployments []map[string]any `json:"deployments"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Deployments, 1)
	strategy.AssertExpectations(t)
}

func TestBacktestRequiresSymbol(t *testing.T) {
	cfg := strategyConfig(new(mocks.MockStrategyClient))
	body, _ := json.Marshal(map[string]any{"config_json": map[string]any{"k": 1}})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/strategies/backtest", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.BacktestStrategy(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestBacktestSuccessAppliesDefaults(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("RunBacktest", mock.Anything, mock.MatchedBy(func(req *strategypb.BacktestRequest) bool {
		return req.Timeframe == "DAY" && req.InitialCapitalCents == 10_000_000 && req.Start != "" && req.End != ""
	})).Return(&strategypb.BacktestResponse{NumTrades: 3}, nil)

	cfg := strategyConfig(strategy)
	body, _ := json.Marshal(map[string]any{"config_json": map[string]any{"k": 1}, "symbol": "AAPL"})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/strategies/backtest", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.BacktestStrategy(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	strategy.AssertExpectations(t)
}

func TestStopDeploymentNotFound(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("StopDeployment", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.NotFound, "missing"))

	cfg := strategyConfig(strategy)
	req := withURLParam(handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/strategies/deployments/d1/stop", nil), "user-1"), "id", "d1")
	w := httptest.NewRecorder()

	cfg.StopDeployment(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
