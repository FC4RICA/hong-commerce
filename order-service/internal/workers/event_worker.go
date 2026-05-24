package workers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/FC4RICA/hong-commerce/order-service/internal/services"
	amqp "github.com/rabbitmq/amqp091-go"
)

type EventWorker struct {
	mqChan *amqp.Channel
	svc    services.OrderService
}

func NewEventWorker(mqChan *amqp.Channel, svc services.OrderService) *EventWorker {
	return &EventWorker{
		mqChan: mqChan,
		svc:    svc,
	}
}

type SagaEvent struct {
	OrderID   uuid.UUID `json:"order_id"`
	PaymentID uuid.UUID `json:"payment_id"`
	Reason    string    `json:"reason"`
}

func (w *EventWorker) Start(ctx context.Context) error {
	// Declare exchange
	err := w.mqChan.ExchangeDeclare(
		"order.events", // name
		"topic",        // type
		true,           // durable
		false,          // auto-deleted
		false,          // internal
		false,          // no-wait
		nil,            // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	// Declare and bind queues for each event type
	events := []struct {
		queue      string
		routingKey string
		handler    func(SagaEvent) error
	}{
		{"order.payment.succeeded", "payment.succeeded", w.handlePaymentSucceeded},
		{"order.inventory.reserved", "inventory.reserved", w.handleInventoryReserved},
		{"order.payment.failed", "payment.failed", w.handlePaymentFailed},
		{"order.inventory.failed", "inventory.failed", w.handleInventoryFailed},
	}

	for _, e := range events {
		_, err := w.mqChan.QueueDeclare(e.queue, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", e.queue, err)
		}

		err = w.mqChan.QueueBind(e.queue, e.routingKey, "order.events", false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s: %w", e.queue, err)
		}

		msgs, err := w.mqChan.Consume(e.queue, "", false, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to consume from %s: %w", e.queue, err)
		}

		go func(handler func(SagaEvent) error, queueName string) {
			for d := range msgs {
				var evt SagaEvent
				if err := json.Unmarshal(d.Body, &evt); err != nil {
					log.Printf("failed to unmarshal event from %s: %v", queueName, err)
					d.Nack(false, false)
					continue
				}

				if err := handler(evt); err != nil {
					log.Printf("failed to handle event from %s: %v", queueName, err)
					d.Nack(false, true) // Requeue on error
					continue
				}

				d.Ack(false)
			}
		}(e.handler, e.queue)
	}

	return nil
}

func (w *EventWorker) handlePaymentSucceeded(evt SagaEvent) error {
	return w.svc.HandlePaymentSucceeded(context.Background(), evt.OrderID, evt.PaymentID)
}

func (w *EventWorker) handleInventoryReserved(evt SagaEvent) error {
	return w.svc.HandleInventoryReserved(context.Background(), evt.OrderID)
}

func (w *EventWorker) handlePaymentFailed(evt SagaEvent) error {
	return w.svc.HandlePaymentFailed(context.Background(), evt.OrderID, evt.Reason)
}

func (w *EventWorker) handleInventoryFailed(evt SagaEvent) error {
	return w.svc.HandleInventoryFailed(context.Background(), evt.OrderID, evt.Reason)
}
