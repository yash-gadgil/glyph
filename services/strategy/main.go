package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yash-gadgil/glyph/services/strategy/server"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := server.InitDB()
	defer db.Close()

	grpcServer := server.NewGrpcServer(os.Getenv("STRATEGY_SVC_PORT"), db)
	if err := grpcServer.Run(ctx); err != nil {
		os.Exit(1)
	}

}
