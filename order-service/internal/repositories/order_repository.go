package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetAll(ctx context.Context) ([]models.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error)
	Update(ctx context.Context, order *models.Order) error
	GetExpiredPendingOrders(ctx context.Context, timeout time.Duration) ([]models.Order, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) GetAll(ctx context.Context) ([]models.Order, error) {
	var orders []models.Order
	err := r.db.WithContext(ctx).Preload("Items").Find(&orders).Error
	return orders, err
}

func (r *orderRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Order, error) {
	var order models.Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Update(ctx context.Context, order *models.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

func (r *orderRepository) GetExpiredPendingOrders(ctx context.Context, timeout time.Duration) ([]models.Order, error) {
	var orders []models.Order
	threshold := time.Now().Add(-timeout)
	err := r.db.WithContext(ctx).
		Where("status = ? AND created_at < ?", "PENDING", threshold).
		Find(&orders).Error
	return orders, err
}
