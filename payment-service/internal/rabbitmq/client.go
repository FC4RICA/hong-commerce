package rabbitmq

import (
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

func NewClient(url string) (*Client, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Declare exchange
	err = ch.ExchangeDeclare(
		"payment_status_exchange", // name
		"direct",                  // type
		true,                      // durable
		false,                     // auto-deleted
		false,                     // internal
		false,                     // no-wait
		nil,                       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare incoming queue
	q, err := ch.QueueDeclare(
		"order_to_payment_queue", // name
		true,                     // durable
		false,                    // delete when unused
		false,                    // exclusive
		false,                    // no-wait
		nil,                      // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	// Example binding to order exchange if needed
	err = ch.QueueBind(
		q.Name,
		"order.created",
		"order_exchange", // assuming the order service uses this
		false,
		nil,
	)
	if err != nil {
		log.Printf("Warning: failed to bind queue to order_exchange (it may not exist yet): %v", err)
	}

	return &Client{
		Conn:    conn,
		Channel: ch,
	}, nil
}

func (c *Client) Close() {
	if c.Channel != nil {
		_ = c.Channel.Close()
	}
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
}
