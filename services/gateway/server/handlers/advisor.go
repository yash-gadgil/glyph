package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	advisorpb "github.com/yash-gadgil/glyph/services/gen/golang/advisor"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (cfg *Config) LoadAdvisorRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Post("/chat", cfg.ChatWithAdvisor)
	r.Get("/chat/session", cfg.GetChatSession)
	r.Delete("/chat/session", cfg.ClearChatSession)
	r.Post("/strategy", cfg.StartStrategyGeneration)
	r.Get("/strategy/status", cfg.GetStrategyJob)
}

func (cfg *Config) StartStrategyGeneration(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("start_strategy_generation"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.advisorClient == nil {
		utils.ReturnErrorJSON(w, "Advisor service unavailable", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Symbol   string `json:"symbol"`
		Provider string `json:"provider"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 30*time.Second)
	defer cancel()

	job, err := cfg.advisorClient.StartStrategyGeneration(ctx, &advisorpb.StartStrategyGenerationRequest{UserId: userID, Symbol: body.Symbol, Provider: body.Provider})
	if err != nil {
		log.Error("start_strategy_generation_error", logger.Stage("rpc"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to start strategy generation", http.StatusInternalServerError)
		return
	}

	writeStrategyJob(w, job)
}

func (cfg *Config) GetStrategyJob(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_strategy_job"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.advisorClient == nil {
		utils.ReturnErrorJSON(w, "Advisor service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
	defer cancel()

	job, err := cfg.advisorClient.GetStrategyJob(ctx, &advisorpb.GetStrategyJobRequest{UserId: userID})
	if err != nil {
		log.Error("get_strategy_job_error", logger.Stage("rpc"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to read strategy job", http.StatusInternalServerError)
		return
	}

	writeStrategyJob(w, job)
}

func writeStrategyJob(w http.ResponseWriter, job *advisorpb.StrategyJob) {
	out := map[string]any{
		"state":      job.State,
		"name":       job.Name,
		"rationale":  job.Rationale,
		"error":      job.Error,
		"started_at": job.StartedAt,
		"updated_at": job.UpdatedAt,
	}
	if job.ConfigJson != "" {
		out["config"] = json.RawMessage(job.ConfigJson)
	}
	if job.Backtest != nil {
		out["backtest"] = map[string]any{
			"total_return_pct": job.Backtest.TotalReturnPct,
			"max_drawdown_pct": job.Backtest.MaxDrawdownPct,
			"sharpe":           job.Backtest.Sharpe,
			"win_rate":         job.Backtest.WinRate,
			"profit_factor":    job.Backtest.ProfitFactor,
			"num_trades":       job.Backtest.NumTrades,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (cfg *Config) ChatWithAdvisor(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("chat_with_advisor"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.advisorClient == nil {
		utils.ReturnErrorJSON(w, "Advisor service unavailable", http.StatusServiceUnavailable)
		return
	}

	var body struct {
		Message  string `json:"message"`
		Provider string `json:"provider"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || strings.TrimSpace(body.Message) == "" {
		utils.ReturnErrorJSON(w, "A message is required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ReturnErrorJSON(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	streamCtx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 6*time.Minute)
	defer cancel()

	stream, err := cfg.advisorClient.ChatWithAdvisor(streamCtx, &advisorpb.ChatRequest{
		UserId:   userID,
		Message:  body.Message,
		Provider: body.Provider,
	})
	if err != nil {
		log.Error("chat_start_error", logger.Stage("open_stream"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to start chat", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			if status.Code(err) == codes.Canceled {
				return
			}
			if status.Code(err) == codes.FailedPrecondition {
				fmt.Fprint(w, "event: busy\ndata: a previous message is still being answered\n\n")
				flusher.Flush()
				return
			}
			log.Error("chat_stream_error", logger.Stage("recv"), zap.Error(err))
			fmt.Fprint(w, "event: error\ndata: chat interrupted\n\n")
			flusher.Flush()
			return
		}
		if chunk.Done {
			break
		}
		if chunk.Text != "" {
			if _, werr := fmt.Fprintf(w, "data: %s\n\n", sseEscape(chunk.Text)); werr != nil {
				return
			}
			flusher.Flush()
		}
	}

	fmt.Fprint(w, "event: done\ndata: end\n\n")
	flusher.Flush()
}

func (cfg *Config) GetChatSession(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("get_chat_session"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.advisorClient == nil {
		utils.ReturnErrorJSON(w, "Advisor service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
	defer cancel()

	session, err := cfg.advisorClient.GetChatSession(ctx, &advisorpb.GetChatSessionRequest{UserId: userID})
	if err != nil {
		log.Error("get_chat_session_error", logger.Stage("rpc"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to load chat", http.StatusInternalServerError)
		return
	}

	turns := make([]map[string]string, 0, len(session.Turns))
	for _, t := range session.Turns {
		turns = append(turns, map[string]string{"role": t.Role, "content": t.Content})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"turns":        turns,
		"in_flight":    session.InFlight,
		"partial_text": session.PartialText,
	})
}

func (cfg *Config) ClearChatSession(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("clear_chat_session"),
	)

	userID, ok := r.Context().Value(userIDKey).(string)
	if !ok || userID == "" {
		utils.ReturnErrorJSON(w, "User ID not found in context", http.StatusUnauthorized)
		return
	}
	if cfg.advisorClient == nil {
		utils.ReturnErrorJSON(w, "Advisor service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 10*time.Second)
	defer cancel()

	if _, err := cfg.advisorClient.ClearChatSession(ctx, &advisorpb.GetChatSessionRequest{UserId: userID}); err != nil {
		log.Error("clear_chat_session_error", logger.Stage("rpc"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to clear chat", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func sseEscape(s string) string {
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if r == '\n' {
			out = append(out, '\\', 'n')
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
