package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yash-gadgil/glyph/services/auth/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcServer := server.NewGrpcServer(os.Getenv("AUTH_SVC_PORT"))
	if err := grpcServer.Run(ctx); err != nil {
		os.Exit(1)
	}
}
