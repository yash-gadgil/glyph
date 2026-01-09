package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/yash-gadgil/glyph/services/order/handlers"
	"github.com/yash-gadgil/glyph/services/order/types"
)

const (
	ExchangeName   = "order.events"
	DLXName        = "order.events.dlx"
	FillsQueueName = "order-svc.fills"
	DoneQueueName  = "order-svc.done"
	FillRoutingKey = "order.fill"
	DoneRoutingKey = "order.done"
	dlqSuffix      = ".dlq"
)

func InitRabbitMQ() (*amqp.Connection, *amqp.Channel) {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	conn, err := amqp.Dial(url)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("failed to open rabbitmq channel: %v", err)
	}

	if err := DeclareTopology(ch); err != nil {
		log.Fatalf("failed to declare rabbitmq topology: %v", err)
	}

	log.Println("connected to rabbitmq")
	return conn, ch
}

func DeclareTopology(ch *amqp.Channel) error {
	if err := ch.ExchangeDeclare(ExchangeName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange: %w", err)
	}
	if err := ch.ExchangeDeclare(DLXName, "direct", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare dlx: %w", err)
	}

	for queue, rk := range map[string]string{
		FillsQueueName: FillRoutingKey,
		DoneQueueName:  DoneRoutingKey,
	} {
		dlq := queue + dlqSuffix
		if _, err := ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
			return fmt.Errorf("declare %s: %w", dlq, err)
		}
		if err := ch.QueueBind(dlq, rk, DLXName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", dlq, err)
		}
		if _, err := ch.QueueDeclare(queue, true, false, false, false, amqp.Table{
			"x-dead-letter-exchange": DLXName,
		}); err != nil {
			return fmt.Errorf("declare %s: %w", queue, err)
		}
		if err := ch.QueueBind(queue, rk, ExchangeName, false, nil); err != nil {
			return fmt.Errorf("bind %s: %w", queue, err)
		}
	}
	return nil
}

type permanentError struct{ error }

func permanent(err error) error { return permanentError{err} }

func isPoison(err error) bool {
	msg := err.Error()
	for _, marker := range []string{"bad trade_id", "bad order_id", "invalid fill qty", "unknown done reason", "fill for unknown order"} {
		if strings.HasPrefix(msg, marker) {
			return true
		}
	}
	return false
}

func ConsumeOrderEvents(ctx context.Context, ch *amqp.Channel, h *handlers.OrderHandler) error {
	fills, err := ch.Consume(FillsQueueName, "order-svc-fills", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume fills: %w", err)
	}
	done, err := ch.Consume(DoneQueueName, "order-svc-done", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume done: %w", err)
	}

	go consumeLoop(ctx, fills, func(body []byte) error {
		var event types.FillEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return permanent(fmt.Errorf("bad fill payload: %w", err))
		}
		if err := h.ApplyFillEvent(ctx, event); err != nil {
			if isPoison(err) {
				return permanent(err)
			}
			return err
		}
		return nil
	})

	go consumeLoop(ctx, done, func(body []byte) error {
		var event types.DoneEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return permanent(fmt.Errorf("bad done payload: %w", err))
		}
		if err := h.ApplyDoneEvent(ctx, event); err != nil {
			if isPoison(err) {
				return permanent(err)
			}
			return err
		}
		return nil
	})

	log.Println("order event consumer started")
	return nil
}

func consumeLoop(ctx context.Context, msgs <-chan amqp.Delivery, handle func([]byte) error) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-msgs:
			if !ok {
				return
			}
			err := handle(msg.Body)
			switch {
			case err == nil:
				_ = msg.Ack(false)
			case errors.As(err, &permanentError{}):
				log.Printf("dead-lettering event: %v (body: %s)", err, msg.Body)
				_ = msg.Nack(false, false)
			default:
				log.Printf("requeueing event after transient failure: %v", err)
				_ = msg.Nack(false, true)
			}
		}
	}
}
