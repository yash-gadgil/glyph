package telemetry

import (
	"context"
	"net/http"
	"os"
	"time"

	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func MetricsAddr() string {
	if v := os.Getenv("METRICS_PORT"); v != "" {
		return v
	}
	return ":9100"
}

func EnableGRPCHistograms() {
	grpc_prometheus.EnableHandlingTimeHistogram()
}

func ServeMetrics(ctx context.Context, addr string, log *zap.Logger) {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{Addr: addr, Handler: mux}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if log != nil {
		log.Info("metrics_server_listening", zap.String("addr", addr))
	}
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed && log != nil {
		log.Error("metrics_server_error", zap.Error(err))
	}
}
