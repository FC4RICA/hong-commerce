package repositories

import (
	"context"

	"github.com/FC4RICA/hong-commerce/order-service/internal/models"
	"gorm.io/gorm"
)

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) error
	GetAll(ctx context.Context) ([]models.Order, error)
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
    err := r.db.WithContext(ctx).Find(&orders).Error
    return orders, err
}
