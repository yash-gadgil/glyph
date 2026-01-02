package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yash-gadgil/glyph/services/order/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db := server.InitDB()
	defer db.Close()

	rmqConn, rmqCh := server.InitRabbitMQ()
	defer rmqConn.Close()
	defer rmqCh.Close()

	orderbookAddr := os.Getenv("ORDERBK_SVC_PORT")
	if orderbookAddr == "" {
		orderbookAddr = "localhost:50056"
	}

	grpcServer := server.NewGrpcServer(
		os.Getenv("ORDER_SVC_PORT"),
		db,
		rmqCh,
		orderbookAddr,
	)
	if err := grpcServer.Run(ctx); err != nil {
		os.Exit(1)
	}
}
