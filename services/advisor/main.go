package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/advisor/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.New("advisor-service")

	grpcServer := server.NewGrpcServer(os.Getenv("ADVISOR_SVC_PORT"))
	if err := grpcServer.Run(ctx); err != nil {
		log.Error("advisor_server_exited", logger.Action("startup"))
		os.Exit(1)
	}
}
