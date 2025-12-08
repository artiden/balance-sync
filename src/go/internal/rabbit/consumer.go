package rabbit

import (
	"encoding/json"
	"log"

	"goservice/internal/config"
	"goservice/internal/service"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Consumer consumes balance change events from RabbitMQ and passes them to the handler.
type Consumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	queueName string
	handler   service.Handler
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

	return &Consumer{
		conn:      conn,
		ch:        ch,
		queueName: cfg.Queue,
		handler:   h,
	}, nil
}

// Start begins consuming messages and blocks until the delivery channel
// is closed or the connection drops.
func (c *Consumer) Start() error {
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
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	for m := range msgs {
		var e service.Event
		if err := json.Unmarshal(m.Body, &e); err != nil {
			log.Printf("rabbit: bad message: %v", err)
			// Reject and do not requeue
			if errAck := m.Nack(false, false); errAck != nil {
				log.Printf("rabbit: nack error for bad message: %v", errAck)
			}
			continue
		}

		if err := c.handler.Handle(e); err != nil {
			log.Printf("rabbit: handler failed: %v", err)
			// requeue message for another attempt
			if errAck := m.Nack(false, true); errAck != nil {
				log.Printf("rabbit: nack error after handler failure: %v", errAck)
			}
			continue
		}

		if err := m.Ack(false); err != nil {
			log.Printf("rabbit: ack error: %v", err)
		}
	}

	return nil
}
