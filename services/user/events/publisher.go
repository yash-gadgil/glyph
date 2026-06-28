package events

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/pkg/userevents"
)

func DeclareUserEventsTopology(ch *amqp.Channel) error {
	return userevents.DeclareTopology(ch)
}

type Publisher struct {
	ch *amqp.Channel
}

func NewPublisher(ch *amqp.Channel) *Publisher {
	return &Publisher{ch: ch}
}

func (p *Publisher) PublishUserDeleted(ctx context.Context, userID string) error {
	return userevents.PublishDeleted(ctx, p.ch, userID)
}
