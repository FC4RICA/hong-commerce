package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/repositories"
	amqp "github.com/rabbitmq/amqp091-go"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, items []models.OrderItem) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error)
	HandlePaymentSucceeded(ctx context.Context, orderID uuid.UUID, paymentID uuid.UUID) error
	HandleInventoryReserved(ctx context.Context, orderID uuid.UUID) error
	HandlePaymentFailed(ctx context.Context, orderID uuid.UUID, reason string) error
	HandleInventoryFailed(ctx context.Context, orderID uuid.UUID, reason string) error
	ProcessTimeoutOrders(ctx context.Context) error
}

type orderService struct {
	repo        repositories.OrderRepository
	mqChan      *amqp.Channel
	timeoutMins int
}

func NewOrderService(repo repositories.OrderRepository, mqChan *amqp.Channel, timeoutMins int) OrderService {
	return &orderService{
		repo:        repo,
		mqChan:      mqChan,
		timeoutMins: timeoutMins,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID uuid.UUID, items []models.OrderItem) (*models.Order, error) {
	// 1. Calculate total amount and prepare order
	var totalAmount float64
	for i := range items {
		items[i].Subtotal = float64(items[i].Quantity) * items[i].UnitPrice
		totalAmount += items[i].Subtotal
	}

	order := &models.Order{
		UserID:      userID,
		Items:       items,
		TotalAmount: totalAmount,
		Status:      "PENDING",
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 2. Publish RabbitMQ Event
	eventItems := make([]map[string]interface{}, len(order.Items))
	for i, it := range order.Items {
		eventItems[i] = map[string]interface{}{
			"product_id":   it.ProductID,
			"product_name": it.ProductName,
			"quantity":     it.Quantity,
			"unit_price":   it.UnitPrice,
			"subtotal":     it.Subtotal,
		}
	}

	event := map[string]interface{}{
		"orderID":   order.ID.String(),
		"amount":    order.TotalAmount,
		"currency":  "THB",
		"items":     eventItems,
		"timestamp": time.Now().Format(time.RFC3339),
	}
	body, _ := json.Marshal(event)

	if s.mqChan != nil {
		// Ensure order_exchange exists before publishing
		_ = s.mqChan.ExchangeDeclare(
			"order_exchange", // name
			"direct",         // type
			true,             // durable
			false,            // auto-deleted
			false,            // internal
			false,            // no-wait
			nil,              // arguments
		)

		err := s.mqChan.PublishWithContext(ctx,
			"order_exchange", // exchange
			"order.created",  // routing key
			false,            // mandatory
			false,            // immediate
			amqp.Publishing{
				ContentType: "application/json",
				Body:        body,
			})
		if err != nil {
			fmt.Printf("failed to publish event: %v\n", err)
		}
	} else {
		fmt.Println("rabbitmq channel is nil, skipping event publication")
	}

	return order, nil
}

func (s *orderService) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	return s.repo.GetByID(ctx, orderID)
}

func (s *orderService) HandlePaymentSucceeded(ctx context.Context, orderID uuid.UUID, paymentID uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.PaymentConfirmed = true
	order.PaymentID = &paymentID

	return s.repo.Update(ctx, order)
}

func (s *orderService) HandleInventoryReserved(ctx context.Context, orderID uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.InventoryConfirmed = true

	return s.repo.Update(ctx, order)
}

func (s *orderService) HandlePaymentFailed(ctx context.Context, orderID uuid.UUID, reason string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.Status = "FAILED"
	// Optional: log reason or notify client
	return s.repo.Update(ctx, order)
}

func (s *orderService) HandleInventoryFailed(ctx context.Context, orderID uuid.UUID, reason string) error {
	order, err := s.repo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}

	order.Status = "FAILED"
	// Note: According to spec, this triggers payment.reverse, but that's Payment-service's responsibility.
	// Order-service just waits/handles the final state.
	return s.repo.Update(ctx, order)
}

func (s *orderService) ProcessTimeoutOrders(ctx context.Context) error {
	// Use configurable timeout
	orders, err := s.repo.GetExpiredPendingOrders(ctx, time.Duration(s.timeoutMins)*time.Minute)
	if err != nil {
		return err
	}

	for _, order := range orders {
		order.Status = "CANCELLED"
		if err := s.repo.Update(ctx, &order); err != nil {
			fmt.Printf("failed to cancel order %s: %v\n", order.ID, err)
		}
		// Optional: Publish order.cancelled event if other services need to know
	}

	return nil
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]models.Order, error) {
    return s.repo.GetAll(ctx)
}

func (s *orderService) GetOrdersByUserID(ctx context.Context, userID uuid.UUID) ([]models.Order, error) {
	return s.repo.GetByUserID(ctx, userID)
}
