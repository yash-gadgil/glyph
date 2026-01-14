package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yash-gadgil/glyph/services/mrktdata/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	grpcServer := server.NewGrpcServer(os.Getenv("MRKTDATA_SVC_PORT"))
	if err := grpcServer.Run(ctx); err != nil {
		os.Exit(1)
	}
}
