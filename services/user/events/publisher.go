package events

import (
	"context"
	"encoding/json"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	UserEventsExchange    = "user.events"
	UserEventsDLX         = "user.events.dlx"
	UserDeletedRoutingKey = "user.deleted"
)

type UserDeletedEvent struct {
	UserID string `json:"user_id"`
}

func DeclareUserEventsTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(UserEventsExchange, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := ch.ExchangeDeclare(UserEventsDLX, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}
	return nil
}

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) *Publisher {
	return &Publisher{ch: ch}
}

func (p *Publisher) PublishUserDeleted(ctx context.Context, userID string) error {
	body, err := json.Marshal(UserDeletedEvent{UserID: userID})
	if err != nil {
		return fmt.Errorf("marshal user.deleted: %w", err)
	}
	return p.ch.PublishWithContext(ctx, UserEventsExchange, UserDeletedRoutingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}
