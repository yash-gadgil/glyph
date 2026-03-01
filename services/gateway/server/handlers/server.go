package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/telemetry"
	"go.uber.org/zap"
)

type Server struct {
	mx  *chi.Mux
	cfg *Config

	log *zap.Logger
}

func NewServer(c *Config) *Server {
	r := chi.NewRouter()

	r.Use(CORSMiddleware)
	r.Use(StructuredLogger(c.log))
	r.Use(telemetry.HTTPMiddleware)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(RateLimiterMiddleware)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"service": "gateway",
			"version": "1.0.0",
		})
	})

	return &Server{mx: r, cfg: c, log: c.log}
}

func (s *Server) ServeAtPort(ctx context.Context, port string) error {
	s.mountRoutes()

	go telemetry.ServeMetrics(ctx, telemetry.MetricsAddr(), s.log)

	s.log.Info("starting_http_server", logger.KV("port", port))

	srv := &http.Server{Addr: port, Handler: s.mx}

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.log.Info("shutting_down_http_server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func (s *Server) AddRoute(routeName string, handlerFunc func(chi.Router)) *Server {
	s.mx.Route(routeName, handlerFunc)
	return s
}

func (s *Server) mountRoutes() {
	s.AddRoute("/auth", s.cfg.LoadAuthRoutes)
	s.AddRoute("/account", s.cfg.LoadAccountRoutes)
	s.AddRoute("/portfolio", s.cfg.LoadPortfolioRoutes)
	s.AddRoute("/watchlists", s.cfg.LoadWatchlistRoutes)
	s.AddRoute("/orders", s.cfg.LoadOrderRoutes)
	s.AddRoute("/explore", s.cfg.LoadExploreRoutes)
	s.AddRoute("/strategies", s.cfg.LoadStrategyRoutes)
}
