package rabbit

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"goservice/internal/config"
	"goservice/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer consumes balance change events from RabbitMQ and passes them to the handler.
type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	queueName   string
	consumerTag string
	handler     service.Handler
}

func NewConsumer(cfg *config.Config, h service.Handler) (*Consumer, error) {
	conn, err := amqp.Dial(cfg.RabbitURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, err
	}

	// Fair dispatch: limit the number of unacknowledged messages.
	if err := ch.Qos(50, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Ensure the queue exists. Durable, non-auto-deleted.
	if _, err := ch.QueueDeclare(
		cfg.Queue,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, err
	}

	// Generate consumer tag for monitoring
	hostname, _ := os.Hostname()
	consumerTag := fmt.Sprintf("go-service-%s-%d", hostname, os.Getpid())

	return &Consumer{
		conn:        conn,
		ch:          ch,
		queueName:   cfg.Queue,
		consumerTag: consumerTag,
		handler:     h,
	}, nil
}

// Start begins consuming messages and blocks until the context is cancelled,
// delivery channel is closed, or the connection drops.
func (c *Consumer) Start(ctx context.Context) error {
	defer func() {
		if err := c.ch.Close(); err != nil {
			log.Printf("rabbit: channel close error: %v", err)
		}
		if err := c.conn.Close(); err != nil {
			log.Printf("rabbit: connection close error: %v", err)
		}
	}()

	msgs, err := c.ch.Consume(
		c.queueName,
		c.consumerTag,
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	log.Printf("rabbit: consumer started with tag %s", c.consumerTag)

	for {
		select {
		case <-ctx.Done():
			log.Printf("rabbit: shutting down consumer")
			return ctx.Err()
		case m, ok := <-msgs:
			if !ok {
				log.Printf("rabbit: message channel closed")
				return nil
			}
			c.processMessage(m)
		}
	}
}

// processMessage handles a single message from the queue.
func (c *Consumer) processMessage(m amqp.Delivery) {
	var e service.Event
	if err := json.Unmarshal(m.Body, &e); err != nil {
		log.Printf("rabbit: bad message format: %v", err)
		// Reject and do not requeue
		if errAck := m.Nack(false, false); errAck != nil {
			log.Printf("rabbit: nack error for bad message: %v", errAck)
		}
		return
	}

	// Validate event fields
	if err := e.Validate(); err != nil {
		log.Printf("rabbit: invalid event: %v", err)
		// Reject and do not requeue - invalid events won't become valid
		if errAck := m.Nack(false, false); errAck != nil {
			log.Printf("rabbit: nack error for invalid event: %v", errAck)
		}
		return
	}

	if err := c.handler.Handle(e); err != nil {
		log.Printf("rabbit: handler failed for event %s: %v", e.EventID, err)
		// requeue message for another attempt
		if errAck := m.Nack(false, true); errAck != nil {
			log.Printf("rabbit: nack error after handler failure: %v", errAck)
		}
		return
	}

	if err := m.Ack(false); err != nil {
		log.Printf("rabbit: ack error: %v", err)
	}
}

// Close gracefully closes the consumer connection.
func (c *Consumer) Close() error {
	if err := c.ch.Cancel(c.consumerTag, false); err != nil {
		return fmt.Errorf("cancelling consumer: %w", err)
	}
	return nil
}
