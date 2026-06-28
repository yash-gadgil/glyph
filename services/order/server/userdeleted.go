package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/logger"
	"github.com/yash-gadgil/glyph/pkg/userevents"
	db "github.com/yash-gadgil/glyph/services/order/db/gen"
	"go.uber.org/zap"
)

const userDeletedConsumerTag = "order-svc-user-deleted"

type userDeletedPermanentError struct{ error }

func userDeletedPermanent(err error) error { return userDeletedPermanentError{err} }

func StartUserDeletedConsumer(ctx context.Context, ch *amqp.Channel, sdb *sql.DB, log *zap.Logger) error {
	if err := userevents.DeclareTopology(ch); err != nil {
		return err
	}

	deliveries, err := ch.Consume(userevents.OrderQueue, userDeletedConsumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume %s: %w", userevents.OrderQueue, err)
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
				err := handleUserDeleted(ctx, sdb, log, msg.Body)
				switch {
				case err == nil:
					_ = msg.Ack(false)
				case errors.As(err, &userDeletedPermanentError{}):
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

func handleUserDeleted(ctx context.Context, sdb *sql.DB, log *zap.Logger, body []byte) error {
	var event userevents.DeletedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return userDeletedPermanent(fmt.Errorf("bad user.deleted payload: %w", err))
	}

	userUUID, err := uuid.Parse(event.UserID)
	if err != nil {
		return userDeletedPermanent(fmt.Errorf("bad user_id %q: %w", event.UserID, err))
	}

	tx, err := sdb.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback()

	qtx := db.New(sdb).WithTx(tx)

	fills, err := qtx.DeleteFillsForUser(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("delete fills for user: %w", err)
	}
	orders, err := qtx.DeleteOrdersForUser(ctx, userUUID)
	if err != nil {
		return fmt.Errorf("delete orders for user: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}

	log.Info("orders_deleted_for_user",
		logger.KV("user_id", event.UserID),
		zap.Int64("orders", orders),
		zap.Int64("fills", fills),
	)
	return nil
}
