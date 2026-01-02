package server

import (
	"fmt"
	"log"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
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
