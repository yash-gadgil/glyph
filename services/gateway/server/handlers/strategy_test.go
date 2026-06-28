package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
	"github.com/yash-gadgil/glyph/services/gateway/tests/mocks"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func strategyConfig(strategy *mocks.MockStrategyClient) *handlers.Config {
	return handlers.NewTestConfig(new(mocks.MockAuthClient)).WithStrategyClient(strategy)
}

func withURLParam(r *http.Request, key, val string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, val)
	return r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))
}

func TestGetStrategiesReturnsList(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("GetStrategies", mock.Anything, &strategypb.UserSpecifier{UserId: "user-1"}).
		Return(&strategypb.StrategiesResponse{
			Strategies: []*strategypb.Strategy{{Id: "s1", Name: "SMA cross"}},
		}, nil)

	cfg := strategyConfig(strategy)
	req := handlers.WithUserID(httptest.NewRequest(http.MethodGet, "/strategies", nil), "user-1")
	w := httptest.NewRecorder()

	cfg.GetStrategies(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body struct {
		Strategies []map[string]any `json:"strategies"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	require.Len(t, body.Strategies, 1)
	assert.Equal(t, "SMA cross", body.Strategies[0]["name"])
	strategy.AssertExpectations(t)
}

func TestCreateStrategyRequiresConfig(t *testing.T) {
	cfg := strategyConfig(new(mocks.MockStrategyClient))

	body, _ := json.Marshal(map[string]any{"name": "Empty"})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/strategies", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.CreateStrategy(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateStrategySuccess(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("CreateStrategy", mock.Anything, mock.Anything).
		Return(&strategypb.Strategy{Id: "s1", Name: "SMA"}, nil)

	cfg := strategyConfig(strategy)
	body, _ := json.Marshal(map[string]any{"name": "SMA", "config_json": map[string]any{"k": 1}})
	req := handlers.WithUserID(httptest.NewRequest(http.MethodPost, "/strategies", bytes.NewReader(body)), "user-1")
	w := httptest.NewRecorder()

	cfg.CreateStrategy(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	strategy.AssertExpectations(t)
}

func TestDeleteStrategyMapsNotFound(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("DeleteStrategy", mock.Anything, mock.Anything).
		Return(nil, status.Error(codes.NotFound, "missing"))

	cfg := strategyConfig(strategy)
	req := withURLParam(handlers.WithUserID(httptest.NewRequest(http.MethodDelete, "/strategies/s1", nil), "user-1"), "id", "s1")
	w := httptest.NewRecorder()

	cfg.DeleteStrategy(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteStrategySuccess(t *testing.T) {
	strategy := new(mocks.MockStrategyClient)
	strategy.On("DeleteStrategy", mock.Anything, mock.Anything).Return(&emptypb.Empty{}, nil)

	cfg := strategyConfig(strategy)
	req := withURLParam(handlers.WithUserID(httptest.NewRequest(http.MethodDelete, "/strategies/s1", nil), "user-1"), "id", "s1")
	w := httptest.NewRecorder()

	cfg.DeleteStrategy(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	strategy.AssertExpectations(t)
}
