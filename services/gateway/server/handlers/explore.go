package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/gateway/server/utils"
	mrktpb "github.com/yash-gadgil/glyph/services/gen/golang/mrktdata"
	"go.uber.org/zap"
)

func (cfg *Config) LoadExploreRoutes(r chi.Router) {
	r.Use(cfg.AuthMiddleware)

	r.Get("/news", cfg.GetNews)
	r.Get("/movers", cfg.GetMovers)
}

func (cfg *Config) GetNews(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_news"))

	var symbols []string
	if raw := r.URL.Query().Get("symbols"); raw != "" {
		symbols = strings.Split(raw, ",")
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if cfg.mrktdataClient == nil {
		utils.ReturnErrorJSON(w, "Market data service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := cfg.mrktdataClient.GetNews(ctx, &mrktpb.NewsRequest{
		Symbols: symbols,
		Limit:   int32(limit),
	})
	if err != nil {
		log.Error("news_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch news", http.StatusInternalServerError)
		return
	}

	type articleJSON struct {
		ID        string   `json:"id"`
		Headline  string   `json:"headline"`
		Summary   string   `json:"summary"`
		Source    string   `json:"source"`
		URL       string   `json:"url"`
		Symbols   []string `json:"symbols"`
		ImageURL  string   `json:"image_url"`
		CreatedAt string   `json:"created_at"`
	}

	articles := make([]articleJSON, len(resp.Articles))
	for i, a := range resp.Articles {
		articles[i] = articleJSON{
			ID:        a.Id,
			Headline:  a.Headline,
			Summary:   a.Summary,
			Source:    a.Source,
			URL:       a.Url,
			Symbols:   a.Symbols,
			ImageURL:  a.ImageUrl,
			CreatedAt: a.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"articles": articles})
}

func (cfg *Config) GetMovers(w http.ResponseWriter, r *http.Request) {
	log := logger.WithContextFields(r.Context(), cfg.log).With(logger.Action("get_movers"))

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if cfg.mrktdataClient == nil {
		utils.ReturnErrorJSON(w, "Market data service unavailable", http.StatusServiceUnavailable)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	resp, err := cfg.mrktdataClient.GetMovers(ctx, &mrktpb.MoversRequest{Limit: int32(limit)})
	if err != nil {
		log.Error("movers_fetch_error", zap.Error(err))
		utils.ReturnErrorJSON(w, "Unable to fetch movers", http.StatusInternalServerError)
		return
	}

	type moverJSON struct {
		Symbol        string  `json:"symbol"`
		CompanyName   string  `json:"company_name"`
		PriceCents    int64   `json:"price_cents"`
		ChangePercent float64 `json:"change_percent"`
		Volume        int64   `json:"volume"`
	}

	convert := func(in []*mrktpb.Mover) []moverJSON {
		out := make([]moverJSON, len(in))
		for i, m := range in {
			out[i] = moverJSON{
				Symbol:        m.Symbol,
				CompanyName:   m.CompanyName,
				PriceCents:    m.PriceCents,
				ChangePercent: m.ChangePercent,
				Volume:        m.Volume,
			}
		}
		return out
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"gainers": convert(resp.Gainers),
		"losers":  convert(resp.Losers),
	})
}
