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

type PaymentCompletedEvent struct {
	PaymentID string `json:"paymentID"`
	OrderID   string `json:"orderID"`
	Status    string `json:"status"`
	PaidAt    string `json:"paidAt"`
}

func (w *EventWorker) Start(ctx context.Context) error {
	// Declare the different exchanges used in the dev stack
	exchanges := []struct {
		name string
		kind string
	}{
		{"order.events", "topic"},
		{"payment_status_exchange", "direct"},
	}

	for _, ex := range exchanges {
		err := w.mqChan.ExchangeDeclare(
			ex.name, // name
			ex.kind, // type
			true,    // durable
			false,   // auto-deleted
			false,   // internal
			false,   // no-wait
			nil,     // arguments
		)
		if err != nil {
			return fmt.Errorf("failed to declare exchange %s: %w", ex.name, err)
		}
	}

	// Declare and bind queues for each event type
	events := []struct {
		queue      string
		exchange   string
		routingKey string
		handler    func([]byte) error
	}{
		{"order.payment.succeeded", "payment_status_exchange", "payment.success", w.handlePaymentSucceeded},
		{"order.inventory.reserved", "order.events", "inventory.reserved", w.handleInventoryReserved},
		{"order.payment.failed", "order.events", "payment.failed", w.handlePaymentFailed},
		{"order.inventory.failed", "order.events", "inventory.failed", w.handleInventoryFailed},
	}

	for _, e := range events {
		_, err := w.mqChan.QueueDeclare(e.queue, true, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to declare queue %s: %w", e.queue, err)
		}

		err = w.mqChan.QueueBind(e.queue, e.routingKey, e.exchange, false, nil)
		if err != nil {
			return fmt.Errorf("failed to bind queue %s to exchange %s: %w", e.queue, e.exchange, err)
		}

		msgs, err := w.mqChan.Consume(e.queue, "", false, false, false, false, nil)
		if err != nil {
			return fmt.Errorf("failed to consume from %s: %w", e.queue, err)
		}

		go func(handler func([]byte) error, queueName string) {
			for d := range msgs {
				if err := handler(d.Body); err != nil {
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

func (w *EventWorker) handlePaymentSucceeded(body []byte) error {
	var evt PaymentCompletedEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment succeeded event: %w", err)
	}

	orderUUID, err := uuid.Parse(evt.OrderID)
	if err != nil {
		return fmt.Errorf("invalid orderID format: %w", err)
	}

	paymentUUID, err := uuid.Parse(evt.PaymentID)
	if err != nil {
		return fmt.Errorf("invalid paymentID format: %w", err)
	}

	return w.svc.HandlePaymentSucceeded(context.Background(), orderUUID, paymentUUID)
}

func (w *EventWorker) handleInventoryReserved(body []byte) error {
	var evt SagaEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal inventory reserved event: %w", err)
	}
	return w.svc.HandleInventoryReserved(context.Background(), evt.OrderID)
}

func (w *EventWorker) handlePaymentFailed(body []byte) error {
	var evt SagaEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal payment failed event: %w", err)
	}
	return w.svc.HandlePaymentFailed(context.Background(), evt.OrderID, evt.Reason)
}

func (w *EventWorker) handleInventoryFailed(body []byte) error {
	var evt SagaEvent
	if err := json.Unmarshal(body, &evt); err != nil {
		return fmt.Errorf("failed to unmarshal inventory failed event: %w", err)
	}
	return w.svc.HandleInventoryFailed(context.Background(), evt.OrderID, evt.Reason)
}
