package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	ordrpb "github.com/yash-gadgil/glyph/services/gen/golang/order"
	strategypb "github.com/yash-gadgil/glyph/services/gen/golang/strategy"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type deploymentJSON struct {
	ID                string `json:"id"`
	StrategyID        string `json:"strategy_id"`
	Symbol            string `json:"symbol"`
	PositionSizeCents int64  `json:"position_size_cents"`
	Status            string `json:"status"`
	InPosition        bool   `json:"in_position"`
	EntryPriceCents   int64  `json:"entry_price_cents"`
	Qty               int64  `json:"qty"`
	StrategyName      string `json:"strategy_name"`
	CreatedAt         string `json:"created_at"`
	UpdatedAt         string `json:"updated_at"`
}

func deploymentToJSON(d *strategypb.Deployment) deploymentJSON {
	return deploymentJSON{
		ID:                d.Id,
		StrategyID:        d.StrategyId,
		Symbol:            d.Symbol,
		PositionSizeCents: d.PositionSizeCents,
		Status:            d.Status,
		InPosition:        d.InPosition,
		EntryPriceCents:   d.EntryPriceCents,
		Qty:               d.Qty,
		StrategyName:      d.StrategyName,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

func (cfg *Config) DeployStrategy(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("deploy_strategy"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}

	strategyID := chi.URLParam(r, "id")
	var body struct {
		Symbol            string `json:"symbol"`
		PositionSizeCents int64  `json:"position_size_cents"`
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

	resp, err := cfg.strategyClient.DeployStrategy(ctx, &strategypb.DeployStrategyRequest{
		StrategyId:        strategyID,
		UserId:            userID,
		Symbol:            body.Symbol,
		PositionSizeCents: body.PositionSizeCents,
	})
	if err != nil {
		log.Error("strategy_deploy_error", zap.Error(err))
		msg := "Unable to deploy strategy"
		if st, ok := status.FromError(err); ok && st.Code() != codes.Internal {
			msg = st.Message()
		}
		utils.ReturnErrorJSON(w, msg, statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(deploymentToJSON(resp))
}

func (cfg *Config) StopDeployment(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("stop_deployment"))

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

	resp, err := cfg.strategyClient.StopDeployment(ctx, &strategypb.DeploymentSpecifier{
		Id:     chi.URLParam(r, "id"),
		UserId: userID,
	})
	if err != nil {
		log.Error("deployment_stop_error", zap.Error(err))
		msg := "Unable to stop deployment"
		if st, ok := status.FromError(err); ok && st.Code() != codes.Internal {
			msg = st.Message()
		}
		utils.ReturnErrorJSON(w, msg, statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deploymentToJSON(resp))
}

func (cfg *Config) DeleteDeployment(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("delete_deployment"))

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

	if _, err := cfg.strategyClient.DeleteDeployment(ctx, &strategypb.DeploymentSpecifier{
		Id:     chi.URLParam(r, "id"),
		UserId: userID,
	}); err != nil {
		log.Error("deployment_delete_error", zap.Error(err))
		msg := "Unable to remove deployment"
		if st, ok := status.FromError(err); ok && st.Code() != codes.Internal {
			msg = st.Message()
		}
		utils.ReturnErrorJSON(w, msg, statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"success": true})
}

func (cfg *Config) GetDeployments(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_deployments"))

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

	resp, err := cfg.strategyClient.GetDeployments(ctx, &strategypb.UserSpecifier{UserId: userID})
	if err != nil {
		log.Error("deployments_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch deployments", http.StatusInternalServerError)
		return
	}

	deployments := make([]deploymentJSON, len(resp.Deployments))
	for i, d := range resp.Deployments {
		deployments[i] = deploymentToJSON(d)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"deployments": deployments})
}

func defaultBacktestWindow(tf string) time.Duration {
	switch tf {
	case "MIN":
		return 7 * 24 * time.Hour
	case "HOUR":
		return 30 * 24 * time.Hour
	default:
		return 182 * 24 * time.Hour
	}
}

func backtestToJSON(resp *strategypb.BacktestResponse) map[string]any {
	type eqJSON struct {
		TimeUnix    int64 `json:"time_unix"`
		EquityCents int64 `json:"equity_cents"`
	}
	type tradeJSON struct {
		EntryTimeUnix   int64   `json:"entry_time_unix"`
		ExitTimeUnix    int64   `json:"exit_time_unix"`
		EntryPriceCents int64   `json:"entry_price_cents"`
		ExitPriceCents  int64   `json:"exit_price_cents"`
		Qty             int64   `json:"qty"`
		PnlCents        int64   `json:"pnl_cents"`
		ReturnPct       float64 `json:"return_pct"`
		HoldBars        int32   `json:"hold_bars"`
		ExitReason      string  `json:"exit_reason"`
	}

	curve := make([]eqJSON, len(resp.EquityCurve))
	for i, p := range resp.EquityCurve {
		curve[i] = eqJSON{TimeUnix: p.TimeUnix, EquityCents: p.EquityCents}
	}
	trades := make([]tradeJSON, len(resp.Trades))
	for i, t := range resp.Trades {
		trades[i] = tradeJSON{
			EntryTimeUnix:   t.EntryTimeUnix,
			ExitTimeUnix:    t.ExitTimeUnix,
			EntryPriceCents: t.EntryPriceCents,
			ExitPriceCents:  t.ExitPriceCents,
			Qty:             t.Qty,
			PnlCents:        t.PnlCents,
			ReturnPct:       t.ReturnPct,
			HoldBars:        t.HoldBars,
			ExitReason:      t.ExitReason,
		}
	}

	return map[string]any{
		"total_return_pct":   resp.TotalReturnPct,
		"max_drawdown_pct":   resp.MaxDrawdownPct,
		"sharpe":             resp.Sharpe,
		"win_rate":           resp.WinRate,
		"profit_factor":      resp.ProfitFactor,
		"num_trades":         resp.NumTrades,
		"avg_hold_bars":      resp.AvgHoldBars,
		"final_equity_cents": resp.FinalEquityCents,
		"bars_used":          resp.BarsUsed,
		"warmup_bars":        resp.WarmupBars,
		"equity_curve":       curve,
		"trades":             trades,
	}
}

func (cfg *Config) BacktestStrategy(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("backtest_strategy"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.strategyClient == nil {
		utils.ReturnErrorJSON(w, "Strategy service unavailable", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		ConfigJSON          json.RawMessage `json:"config_json"`
		Symbol              string          `json:"symbol"`
		Timeframe           string          `json:"timeframe"`
		Start               string          `json:"start"`
		End                 string          `json:"end"`
		InitialCapitalCents int64           `json:"initial_capital_cents"`
		PositionSizeCents   int64           `json:"position_size_cents"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.ReturnErrorJSON(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if len(body.ConfigJSON) == 0 {
		utils.ReturnErrorJSON(w, "config_json is required", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(body.Symbol) == "" {
		utils.ReturnErrorJSON(w, "symbol is required", http.StatusBadRequest)
		return
	}

	timeframe := strings.ToUpper(strings.TrimSpace(body.Timeframe))
	if timeframe == "" {
		timeframe = "DAY"
	}
	end := strings.TrimSpace(body.End)
	if end == "" {
		end = time.Now().UTC().Format("2006-01-02")
	}
	start := strings.TrimSpace(body.Start)
	if start == "" {
		endT, err := time.Parse("2006-01-02", end)
		if err != nil {
			endT = time.Now().UTC()
		}
		start = endT.Add(-defaultBacktestWindow(timeframe)).Format("2006-01-02")
	}
	if body.InitialCapitalCents <= 0 {
		body.InitialCapitalCents = 10_000_000
	}
	if body.PositionSizeCents <= 0 {
		body.PositionSizeCents = 1_000_000
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	resp, err := cfg.strategyClient.RunBacktest(ctx, &strategypb.BacktestRequest{
		UserId:              userID,
		ConfigJson:          string(body.ConfigJSON),
		Symbol:              body.Symbol,
		Timeframe:           timeframe,
		Start:               start,
		End:                 end,
		InitialCapitalCents: body.InitialCapitalCents,
		PositionSizeCents:   body.PositionSizeCents,
	})
	if err != nil {
		log.Error("backtest_error", zap.Error(err))
		msg := "Unable to run backtest"
		if st, ok := status.FromError(err); ok && st.Code() != codes.Internal {
			msg = st.Message()
		}
		utils.ReturnErrorJSON(w, msg, statusFromGrpc(err, http.StatusInternalServerError))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(backtestToJSON(resp))
}

func (cfg *Config) GetStrategyTrades(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_strategy_trades"))

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.orderClient == nil {
		utils.ReturnErrorJSON(w, "Order service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	resp, err := cfg.orderClient.GetStrategyFills(ctx, &ordrpb.GetStrategyFillsRequest{
		StrategyId: chi.URLParam(r, "id"),
		UserId:     userID,
		Limit:      100,
	})
	if err != nil {
		log.Error("strategy_trades_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch strategy trades", http.StatusInternalServerError)
		return
	}

	type fillJSON struct {
		TradeID    string `json:"trade_id"`
		OrderID    string `json:"order_id"`
		Symbol     string `json:"symbol"`
		Side       string `json:"side"`
		Qty        int64  `json:"qty"`
		PriceCents int64  `json:"price_cents"`
		ExecutedAt string `json:"executed_at"`
	}
	fills := make([]fillJSON, len(resp.Fills))
	for i, f := range resp.Fills {
		side := "buy"
		if f.Side == ordrpb.Side_SELL {
			side = "sell"
		}
		executedAt := ""
		if f.ExecutedAt != nil {
			executedAt = f.ExecutedAt.AsTime().Format(time.RFC3339)
		}
		fills[i] = fillJSON{
			TradeID:    f.TradeId,
			OrderID:    f.OrderId,
			Symbol:     f.Symbol,
			Side:       side,
			Qty:        f.Qty,
			PriceCents: f.PriceCents,
			ExecutedAt: executedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"fills": fills})
}
