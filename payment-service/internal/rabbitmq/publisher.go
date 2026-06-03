package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/services"
	amqp "github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	client *Client
}

func NewPublisher(client *Client) *Publisher {
	return &Publisher{
		client: client,
	}
}

func (p *Publisher) PublishPaymentCompleted(ctx context.Context, event services.PaymentCompletedEvent) error {
	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	err = p.client.Channel.PublishWithContext(ctx,
		"payment_status_exchange", // exchange
		"payment.success",         // routing key
		false,                     // mandatory
		false,                     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
			Timestamp:   time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}
