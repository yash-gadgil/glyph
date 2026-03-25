package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

	r.Get("/analyze", cfg.AnalyzePortfolio)
	r.Post("/strategy", cfg.GenerateStrategy)
}

func (cfg *Config) GenerateStrategy(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("generate_strategy"),
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

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 3*time.Minute)
	defer cancel()

	res, err := cfg.advisorClient.GenerateStrategy(ctx, &advisorpb.GenerateStrategyRequest{UserId: userID})
	if err != nil {
		log.Error("generate_strategy_error", logger.Stage("rpc"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to generate a strategy", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"name":      res.Name,
		"rationale": res.Rationale,
		"template":  res.Template,
		"config":    json.RawMessage(res.ConfigJson),
	})
}

func (cfg *Config) AnalyzePortfolio(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(
		logger.Action("analyze_portfolio"),
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

	flusher, ok := w.(http.Flusher)
	if !ok {
		utils.ReturnErrorJSON(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	streamCtx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 5*time.Minute)
	defer cancel()

	stream, err := cfg.advisorClient.AnalyzePortfolio(streamCtx, &advisorpb.AnalyzeRequest{
		UserId: userID,
	})
	if err != nil {
		log.Error("analyze_start_error", logger.Stage("open_stream"), zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to start analysis", http.StatusInternalServerError)
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
			log.Error("analyze_stream_error", logger.Stage("recv"), zap.Error(err))
			fmt.Fprint(w, "event: error\ndata: analysis interrupted\n\n")
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
