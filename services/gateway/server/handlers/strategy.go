package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
)

func (cfg *Config) LoadStrategyRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Get("/", cfg.GetStrategies)
	r.Post("/", cfg.CreateStrategy)
	r.Patch("/{id}", cfg.UpdateStrategy)
	r.Delete("/{id}", cfg.DeleteStrategy)
}

type strategyJSON struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	ConfigJSON string `json:"config_json"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}

func strategyToJSON(s *userpb.Strategy) strategyJSON {
	return strategyJSON{
		ID:         s.Id,
		Name:       s.Name,
		ConfigJSON: s.ConfigJson,
		CreatedAt:  s.CreatedAt,
		UpdatedAt:  s.UpdatedAt,
	}
}

func (cfg *Config) GetStrategies(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_strategies"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if cfg.strategyClient == nil {
		utils.ReturnErrorJSON(w, "Strategy service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := cfg.strategyClient.GetStrategies(ctx, &userpb.UserSpecifier{UserId: userID})
	if err != nil {
		log.Error("strategies_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch strategies", http.StatusInternalServerError)
		return
	}

	strategies := make([]strategyJSON, len(resp.Strategies))
	for i, s := range resp.Strategies {
		strategies[i] = strategyToJSON(s)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"strategies": strategies})
}

func (cfg *Config) CreateStrategy(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("create_strategy"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	var body struct {
		Name       string          `json:"name"`
		ConfigJSON json.RawMessage `json:"config_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.ReturnErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if body.Name == "" || len(body.ConfigJSON) == 0 {
		utils.ReturnErrorJSON(w, "name and config_json are required", http.StatusBadRequest)
		return
	}

	if cfg.strategyClient == nil {
		utils.ReturnErrorJSON(w, "Strategy service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := cfg.strategyClient.CreateStrategy(ctx, &userpb.CreateStrategyRequest{
		UserId:     userID,
		Name:       body.Name,
		ConfigJson: string(body.ConfigJSON),
	})
	if err != nil {
		log.Error("strategy_create_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to create strategy", statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(strategyToJSON(resp))
}

func (cfg *Config) UpdateStrategy(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("update_strategy"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	strategyID := chi.URLParam(r, "id")
	if strategyID == "" {
		utils.ReturnErrorJSON(w, "Strategy ID is required", http.StatusBadRequest)
		return
	}

	var body struct {
		Name       string          `json:"name"`
		ConfigJSON json.RawMessage `json:"config_json"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.ReturnErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if cfg.strategyClient == nil {
		utils.ReturnErrorJSON(w, "Strategy service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := cfg.strategyClient.UpdateStrategy(ctx, &userpb.UpdateStrategyRequest{
		Id:         strategyID,
		UserId:     userID,
		Name:       body.Name,
		ConfigJson: string(body.ConfigJSON),
	})
	if err != nil {
		log.Error("strategy_update_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to update strategy", statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(strategyToJSON(resp))
}

func (cfg *Config) DeleteStrategy(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("delete_strategy"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	strategyID := chi.URLParam(r, "id")
	if strategyID == "" {
		utils.ReturnErrorJSON(w, "Strategy ID is required", http.StatusBadRequest)
		return
	}

	if cfg.strategyClient == nil {
		utils.ReturnErrorJSON(w, "Strategy service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	if _, err := cfg.strategyClient.DeleteStrategy(ctx, &userpb.StrategySpecifier{
		Id:     strategyID,
		UserId: userID,
	}); err != nil {
		log.Error("strategy_delete_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to delete strategy", statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}
