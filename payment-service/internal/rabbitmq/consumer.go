package rabbitmq

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/services"
)

type Consumer struct {
	client  *Client
	service services.PaymentService
}

func NewConsumer(client *Client, service services.PaymentService) *Consumer {
	return &Consumer{
		client:  client,
		service: service,
	}
}

type OrderCreatedEvent struct {
	OrderID   string  `json:"orderID"`
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Timestamp string  `json:"timestamp"`
}

func (c *Consumer) Start() error {
	msgs, err := c.client.Channel.Consume(
		"order_to_payment_queue", // queue
		"",                       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	go func() {
		for d := range msgs {
			var event OrderCreatedEvent
			if err := json.Unmarshal(d.Body, &event); err != nil {
				log.Printf("Error decoding message: %v", err)
				_ = d.Nack(false, false) // send to DLQ ideally
				continue
			}

			err := c.service.InitializePayment(event.OrderID, event.Currency, event.Amount)
			if err != nil {
				log.Printf("Failed to process order %s: %v", event.OrderID, err)
				_ = d.Nack(false, true) // requeue
			} else {
				log.Printf("Successfully processed payment creation for order %s", event.OrderID)
				_ = d.Ack(false)
			}
		}
	}()

	return nil
}
