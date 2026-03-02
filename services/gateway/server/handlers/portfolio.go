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
	userpb "github.com/yash-gadgil/glyph/services/gen/golang/user"
	"go.uber.org/zap"
)

func (cfg *Config) LoadPortfolioRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Get("/", cfg.GetPortfolio)
	r.Get("/holdings", cfg.GetHoldings)
	r.Get("/positions", cfg.GetPositions)
	r.Get("/history", cfg.GetPortfolioHistory)
}

func (cfg *Config) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_portfolio"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.portfolioClient == nil {
		utils.ReturnErrorJSON(w, "Portfolio service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	res, err := cfg.portfolioClient.GetPortfolio(ctx, &userpb.UserSpecifier{
		UserId: userID,
	})
	if err != nil {
		log.Error("portfolio_fetch_error", logger.Stage("fetch_portfolio"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch portfolio", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"cash_balance_cents":  res.CashBalanceCents,
		"reserved_cash_cents": res.ReservedCashCents,
		"buying_power_cents":  res.CashBalanceCents - res.ReservedCashCents,
		"currency":            res.Currency,
		"multiplier":          res.Multiplier,
		"margin_used_cents":   res.MarginUsedCents,
	})
}

func (cfg *Config) GetHoldings(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_holdings"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.portfolioClient == nil {
		utils.ReturnErrorJSON(w, "Portfolio service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	res, err := cfg.portfolioClient.GetHoldings(ctx, &userpb.UserSpecifier{UserId: userID})
	if err != nil {
		log.Error("holdings_fetch_error", logger.Stage("fetch_holdings"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch holdings", http.StatusInternalServerError)
		return
	}

	type holdingJSON struct {
		Symbol             string `json:"symbol"`
		Qty                int64  `json:"qty"`
		AvgPriceCents      int64  `json:"avg_price_cents"`
		CostBasisCents     int64  `json:"cost_basis_cents"`
		LastPriceCents     int64  `json:"last_price_cents"`
		MarketValueCents   int64  `json:"market_value_cents"`
		UnrealizedPnlCents int64  `json:"unrealized_pnl_cents"`
		RealizedPnlCents   int64  `json:"realized_pnl_cents"`
	}

	holdings := make([]holdingJSON, len(res.Holdings))
	for i, h := range res.Holdings {
		holdings[i] = holdingJSON{
			Symbol:             h.Symbol,
			Qty:                h.Qty,
			AvgPriceCents:      h.AvgPriceCents,
			CostBasisCents:     h.CostBasisCents,
			LastPriceCents:     h.LastPriceCents,
			MarketValueCents:   h.MarketValueCents,
			UnrealizedPnlCents: h.UnrealizedPnlCents,
			RealizedPnlCents:   h.RealizedPnlCents,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"holdings":                   holdings,
		"total_market_value_cents":   res.TotalMarketValueCents,
		"total_cost_basis_cents":     res.TotalCostBasisCents,
		"total_unrealized_pnl_cents": res.TotalUnrealizedPnlCents,
		"total_realized_pnl_cents":   res.TotalRealizedPnlCents,
	})
}

func (cfg *Config) GetPositions(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_positions"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.portfolioClient == nil {
		utils.ReturnErrorJSON(w, "Portfolio service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*3)
	defer cancel()

	res, err := cfg.portfolioClient.GetPositions(ctx, &userpb.UserSpecifier{
		UserId: userID,
	})
	if err != nil {
		log.Error("positions_fetch_error", logger.Stage("fetch_positions"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch positions", http.StatusInternalServerError)
		return
	}

	type positionJSON struct {
		Symbol           string `json:"symbol"`
		Qty              int64  `json:"qty"`
		ReservedQty      int64  `json:"reserved_qty"`
		AvgPriceCents    int64  `json:"avg_price_cents"`
		CostBasisCents   int64  `json:"cost_basis_cents"`
		RealizedPnlCents int64  `json:"realized_pnl_cents"`
	}

	positions := make([]positionJSON, len(res.Positions))
	for i, p := range res.Positions {
		positions[i] = positionJSON{
			Symbol:           p.Symbol,
			Qty:              p.Qty,
			ReservedQty:      p.ReservedQty,
			AvgPriceCents:    p.AvgPriceCents,
			CostBasisCents:   p.CostBasisCents,
			RealizedPnlCents: p.RealizedPnlCents,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"positions": positions})
}

func (cfg *Config) GetPortfolioHistory(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_portfolio_history"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		log.Error("user_id_missing", logger.Stage("context_extraction"))
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.portfolioClient == nil {
		utils.ReturnErrorJSON(w, "Portfolio service unavailable", http.StatusServiceUnavailable)
		return
	}

	hours := 24
	if v := r.URL.Query().Get("hours"); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			hours = parsed
		}
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*5)
	defer cancel()

	res, err := cfg.portfolioClient.GetPortfolioHistory(ctx, &userpb.PortfolioHistoryRequest{
		UserId: userID,
		Hours:  int32(hours),
	})
	if err != nil {
		log.Error("history_fetch_error", logger.Stage("fetch_history"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch portfolio history", http.StatusInternalServerError)
		return
	}

	type pointJSON struct {
		TimeUnix         int64 `json:"time_unix"`
		EquityCents      int64 `json:"equity_cents"`
		CashCents        int64 `json:"cash_cents"`
		MarketValueCents int64 `json:"market_value_cents"`
	}

	points := make([]pointJSON, len(res.Points))
	for i, p := range res.Points {
		points[i] = pointJSON{
			TimeUnix:         p.TimeUnix,
			EquityCents:      p.EquityCents,
			CashCents:        p.CashCents,
			MarketValueCents: p.MarketValueCents,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"points": points})
}
