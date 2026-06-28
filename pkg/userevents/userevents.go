package userevents

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeName      = "user.events"
	DLXName           = "user.events.dlx"
	DeletedRoutingKey = "user.deleted"
	StrategyQueue     = "strategy-svc.user-deleted"
	OrderQueue        = "order-svc.user-deleted"
	dlqSuffix         = ".dlq"
)

var consumerQueues = []string{StrategyQueue, OrderQueue}

type DeletedEvent struct {
	UserID string `json:"user_id"`
}

func DeclareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := ch.ExchangeDeclare(DLXName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}
	for _, queue := range consumerQueues {
		dlq := queue + dlqSuffix
		if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare %s: %w", dlq, err)
		}
		if err := ch.QueueBind(dlq, DeletedRoutingKey, DLXName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", dlq, err)
		}
		if _, err := ch.QueueDeclare(queue, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange": DLXName,
		}); err != nil {
			return fmt.Errorf("declare %s: %w", queue, err)
		}
		if err := ch.QueueBind(queue, DeletedRoutingKey, ExchangeName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", queue, err)
		}
	}
	return nil
}

func PublishDeleted(ctx context.Context, ch *amqp.Channel, userID string) error {
	body, err := json.Marshal(DeletedEvent{UserID: userID})
	if err != nil {
		return fmt.Errorf("marshal user.deleted: %w", err)
	}
	return ch.PublishWithContext(ctx, ExchangeName, DeletedRoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}
