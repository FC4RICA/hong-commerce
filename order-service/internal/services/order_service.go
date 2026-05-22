package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"github.com/FC4RICA/hong-commerce/order-service/internal/repositories"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type OrderService interface {
	CreateOrder(ctx context.Context, userID, itemID string, quantity int, totalPrice float64) (*models.Order, error)
	GetAllOrders(ctx context.Context) ([]models.Order, error)
}

type orderService struct {
	repo    repositories.OrderRepository
	redis   *redis.Client
	mqChan  *amqp.Channel
}

func NewOrderService(repo repositories.OrderRepository, rdb *redis.Client, mqChan *amqp.Channel) OrderService {
	return &orderService{
		repo:   repo,
		redis:  rdb,
		mqChan: mqChan,
	}
}

func (s *orderService) CreateOrder(ctx context.Context, userID, itemID string, quantity int, totalPrice float64) (*models.Order, error) {
	// 1. Redis Lock / Idempotency check (Prevent double-click within 5s)
	lockKey := fmt.Sprintf("lock:order:%s:%s", userID, itemID)
	success, err := s.redis.SetNX(ctx, lockKey, "locked", 5*time.Second).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}
	if !success {
		return nil, fmt.Errorf("duplicate order request, please wait")
	}

	// 2. Create Order in Database
	order := &models.Order{
		UserID:     userID,
		ItemID:     itemID,
		Quantity:   quantity,
		TotalPrice: totalPrice,
		Status:     "PENDING",
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// 3. Publish RabbitMQ Event
	event := map[string]interface{}{
		"order_id":    order.ID,
		"user_id":     order.UserID,
		"item_id":     order.ItemID,
		"quantity":    order.Quantity,
		"total_price": order.TotalPrice,
		"status":      order.Status,
		"timestamp":   time.Now().Unix(),
	}
	body, _ := json.Marshal(event)

	err = s.mqChan.PublishWithContext(ctx,
		"",               // exchange
		"order.created",  // routing key (queue name)
		false,            // mandatory
		false,            // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		// In a real scenario, we might want to use a transactional outbox pattern
		// or retry mechanism here if the MQ publish fails.
		fmt.Printf("failed to publish event: %v\n", err)
	}

	return order, nil
}

func (s *orderService) GetAllOrders(ctx context.Context) ([]models.Order, error) {
    return s.repo.GetAll(ctx)
}
