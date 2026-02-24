package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/yash-gadgil/glyph/services/gateway/server/handlers"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := handlers.NewFromEnv()
	defer cfg.Close()

	srv := handlers.NewServer(cfg)

	if err := srv.ServeAtPort(ctx, cfg.GatewayServiceAddr); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("error starting server: ", err)
	}
}
