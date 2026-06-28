package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/services/user/events"
	"github.com/yash-gadgil/glyph/services/user/handlers"
	"github.com/yash-gadgil/glyph/services/user/server"
	"github.com/yash-gadgil/glyph/services/user/worker"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	log := logger.New("user-service")

	db := server.InitDB()
	defer db.Close()

	var pub handlers.UserEventPublisher

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

		if err := worker.StartSettlementConsumer(ctx, ch, worker.NewSettler(db, log), log); err != nil {
			log.Fatal("settlement_consumer_failed", zap.Error(err))
		}

		pubCh, err := conn.Channel()
		if err != nil {
			log.Fatal("rabbitmq_publish_channel_failed", zap.Error(err))
		}
		defer pubCh.Close()

		if err := events.DeclareUserEventsTopology(pubCh); err != nil {
			log.Fatal("user_events_topology_failed", zap.Error(err))
		}
		pub = events.NewPublisher(pubCh)
	}

	grpcServer := server.NewGrpcServer(os.Getenv("USER_SVC_PORT"), db, pub)
	if err := grpcServer.Run(ctx); err != nil {
		os.Exit(1)
	}
}
