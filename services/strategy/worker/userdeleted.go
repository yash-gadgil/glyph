package worker

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/logger"
	db "github.com/yash-gadgil/glyph/services/strategy/db/gen"
	"go.uber.org/zap"
)

const (
	UserEventsExchange    = "user.events"
	UserEventsDLX         = "user.events.dlx"
	UserDeletedRoutingKey = "user.deleted"
	UserDeletedQueueName  = "strategy-svc.user-deleted"
	dlqSuffix             = ".dlq"
	consumerTag           = "strategy-svc-user-deleted"
)

type userDeletedEvent struct {
	UserID string `json:"user_id"`
}

func DeclareUserDeletedTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(UserEventsExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := ch.ExchangeDeclare(UserEventsDLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}

	dlq := UserDeletedQueueName + dlqSuffix
	if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare %s: %w", dlq, err)
	}
	if err := ch.QueueBind(dlq, UserDeletedRoutingKey, UserEventsDLX, false, nil); err != nil {
		return fmt.Errorf("bind %s: %w", dlq, err)
	}
	if _, err := ch.QueueDeclare(UserDeletedQueueName, true, false, false, false, amqp.Table{
		"x-dead-letter-exchange": UserEventsDLX,
	}); err != nil {
		return fmt.Errorf("declare %s: %w", UserDeletedQueueName, err)
	}
	if err := ch.QueueBind(UserDeletedQueueName, UserDeletedRoutingKey, UserEventsExchange, false, nil); err != nil {
		return fmt.Errorf("bind %s: %w", UserDeletedQueueName, err)
	}
	return nil
}

type permanentError struct{ error }

func permanent(err error) error { return permanentError{err} }

func StartUserDeletedConsumer(ctx context.Context, ch *amqp.Channel, q *db.Queries, log *zap.Logger) error {
	if err := DeclareUserDeletedTopology(ch); err != nil {
		return err
	}

	deliveries, err := ch.Consume(UserDeletedQueueName, consumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", UserDeletedQueueName, err)
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
	var event userDeletedEvent
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
