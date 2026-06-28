package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/logger"
	gen "github.com/yash-gadgil/glyph/services/strategy/db/gen"
	"github.com/yash-gadgil/glyph/services/strategy/server"
	"github.com/yash-gadgil/glyph/services/strategy/worker"
	"go.uber.org/zap"
)

func main() {

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.New("strategy-service")

	db := server.InitDB()
	defer db.Close()

	if url := os.Getenv("RABBITMQ_URL"); url != "" {
		conn, err := amqp.Dial(url)
		if err != nil {
			log.Fatal("rabbitmq_dial_failed", zap.Error(err))
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			log.Fatal("rabbitmq_channel_failed", zap.Error(err))
		}
		defer ch.Close()

		if err := worker.StartUserDeletedConsumer(ctx, ch, gen.New(db), log); err != nil {
			log.Fatal("user_deleted_consumer_failed", zap.Error(err))
		}
	}

	grpcServer := server.NewGrpcServer(os.Getenv("STRATEGY_SVC_PORT"), db)
	if err := grpcServer.Run(ctx); err != nil {
		os.Exit(1)
	}

}
