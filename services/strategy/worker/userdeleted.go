package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/userevents"
	db "github.com/yash-gadgil/glyph/services/strategy/db/gen"
	"go.uber.org/zap"
)

const consumerTag = "strategy-svc-user-deleted"

type permanentError struct{ error }

func permanent(err error) error { return permanentError{err} }

func StartUserDeletedConsumer(ctx context.Context, ch *amqp.Channel, q *db.Queries, log *zap.Logger) error {
	if err := userevents.DeclareTopology(ch); err != nil {
		return err
	}

	deliveries, err := ch.Consume(userevents.StrategyQueue, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", userevents.StrategyQueue, err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-deliveries:
				if !ok {
					return
				}
				err := handleUserDeleted(ctx, q, log, msg.Body)
				switch {
				case err == nil:
					_ = msg.Ack(false)
				case errors.As(err, &permanentError{}):
					log.Error("user_deleted_dead_lettered", zap.Error(err), logger.KV("body", string(msg.Body)))
					_ = msg.Nack(false, false)
				default:
					log.Warn("user_deleted_requeued", zap.Error(err))
					_ = msg.Nack(false, true)
				}
			}
		}
	}()

	log.Info("user_deleted_consumer_started")
	return nil
}

func handleUserDeleted(ctx context.Context, q *db.Queries, log *zap.Logger, body []byte) error {
	var event userevents.DeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return permanent(fmt.Errorf("bad user.deleted payload: %w", err))
	}

	userUUID, err := uuid.Parse(event.UserID)
	if err != nil {
		return permanent(fmt.Errorf("bad user_id %q: %w", event.UserID, err))
	}

	deleted, err := q.DeleteStrategiesForUser(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("delete strategies for user: %w", err)
	}

	log.Info("strategies_deleted_for_user", logger.KV("user_id", event.UserID), zap.Int64("count", deleted))
	return nil
}
