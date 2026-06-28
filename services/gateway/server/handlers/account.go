package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
)

func (cfg *Config) LoadAccountRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Get("/", cfg.GetAccount)
	r.Delete("/", cfg.DeleteAccount)
	r.Post("/reset", cfg.ResetAccount)
	r.Get("/profile", cfg.GetProfile)
	r.Get("/trades", cfg.GetTrades)
	r.Get("/me", cfg.Me)
}

func (cfg *Config) GetAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_account"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if cfg.accountClient == nil || cfg.portfolioClient == nil {
		utils.ReturnErrorJSON(w, "Account service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	spec := &userpb.UserSpecifier{UserId: userID}

	profile, err := cfg.accountClient.GetProfile(ctx, spec)
	if err != nil {
		log.Error("profile_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch account", http.StatusInternalServerError)
		return
	}

	portfolio, err := cfg.portfolioClient.GetPortfolio(ctx, spec)
	if err != nil {
		log.Error("portfolio_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch account", http.StatusInternalServerError)
		return
	}

	var holdingsValue int64
	if holdings, err := cfg.portfolioClient.GetHoldings(ctx, spec); err == nil {
		holdingsValue = holdings.TotalMarketValueCents
	} else {
		log.Warn("holdings_fetch_failed_equity_is_cash_only", zap.Error(err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user_id":             profile.UserId,
		"email":               profile.Email,
		"user_name":           profile.UserName,
		"cash_balance_cents":  portfolio.CashBalanceCents,
		"reserved_cash_cents": portfolio.ReservedCashCents,
		"buying_power_cents":  portfolio.CashBalanceCents - portfolio.ReservedCashCents,
		"equity_cents":        portfolio.CashBalanceCents + holdingsValue,
		"currency":            portfolio.Currency,
		"multiplier":          portfolio.Multiplier,
	})
}

func (cfg *Config) ResetAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("reset_account"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if cfg.accountClient == nil {
		utils.ReturnErrorJSON(w, "Account service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := cfg.accountClient.ResetAccount(ctx, &userpb.UserSpecifier{UserId: userID}); err != nil {
		log.Error("reset_account_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to reset account", http.StatusInternalServerError)
		return
	}

	log.Info("account_reset", logger.KV("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (cfg *Config) DeleteAccount(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("delete_account"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if cfg.accountClient == nil {
		utils.ReturnErrorJSON(w, "Account service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	if _, err := cfg.accountClient.DeleteAccount(ctx, &userpb.UserSpecifier{UserId: userID}); err != nil {
		log.Error("delete_account_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to delete account", statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	log.Info("account_deleted", logger.KV("user_id", userID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (cfg *Config) GetProfile(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_profile"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	if cfg.accountClient == nil {
		utils.ReturnErrorJSON(w, "Account service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	profile, err := cfg.accountClient.GetProfile(ctx, &userpb.UserSpecifier{UserId: userID})
	if err != nil {
		log.Error("profile_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"user_id":   profile.UserId,
		"email":     profile.Email,
		"user_name": profile.UserName,
	})
}

func (cfg *Config) GetTrades(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_trades"))

	if cfg.orderClient == nil {
		utils.ReturnErrorJSON(w, "Order service unavailable", http.StatusServiceUnavailable)
		return
	}

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	limit, offset := parsePageParams(r)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := cfg.orderClient.GetFills(ctx, &ordrpb.GetFillsRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		log.Error("fills_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch trades", http.StatusInternalServerError)
		return
	}

	type fillJSON struct {
		TradeID    string `json:"trade_id"`
		OrderID    string `json:"order_id"`
		Symbol     string `json:"symbol"`
		Side       string `json:"side"`
		Qty        int64  `json:"qty"`
		PriceCents int64  `json:"price_cents"`
		Liquidity  string `json:"liquidity"`
		ExecutedAt string `json:"executed_at"`
	}

	fills := make([]fillJSON, len(resp.Fills))
	for i, f := range resp.Fills {
		fills[i] = fillJSON{
			TradeID:    f.TradeId,
			OrderID:    f.OrderId,
			Symbol:     f.Symbol,
			Side:       sideProtoToString[f.Side],
			Qty:        f.Qty,
			PriceCents: f.PriceCents,
			Liquidity:  f.Liquidity,
		}
		if f.ExecutedAt != nil {
			fills[i].ExecutedAt = f.ExecutedAt.AsTime().Format(time.RFC3339)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"fills": fills})
}

func (cfg *Config) Me(w http.ResponseWriter, r *http.Request) {
	res := struct {
		Id string `json:"id"`
	}{
		Id: r.Context().Value(userIDKey).(string),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func parsePageParams(r *http.Request) (int32, int32) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit < 0 {
		limit = 0
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	return int32(limit), int32(offset)
}

var sideProtoToString = map[ordrpb.Side]string{
	ordrpb.Side_BUY:  "buy",
	ordrpb.Side_SELL: "sell",
}
