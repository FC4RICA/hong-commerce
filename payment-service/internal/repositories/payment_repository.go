package repositories

import (
	"errors"

	"github.com/FC4RICA/hong-commerce/payment-service/internal/entities"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	CreatePaymentWithIdempotency(payment *entities.Payment, idempotencyKey string) error
	UpdatePaymentWithIdempotency(payment *entities.Payment, idempotencyKey string) error
	GetPaymentByID(id string) (*entities.Payment, error)
	CheckIdempotencyKey(key string) (bool, error)
}

type paymentRepositoryImpl struct {
	db *gorm.DB
}

func NewPaymentRepository(db *gorm.DB) PaymentRepository {
	return &paymentRepositoryImpl{db: db}
}

func (r *paymentRepositoryImpl) CheckIdempotencyKey(key string) (bool, error) {
	var existingKey entities.IdempotencyKey
	err := r.db.Where("key = ?", key).First(&existingKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil // Not found, so not processed yet
		}
		return false, err // DB error
	}
	return true, nil // Found, already processed
}

func (r *paymentRepositoryImpl) CreatePaymentWithIdempotency(payment *entities.Payment, idempotencyKey string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(payment).Error; err != nil {
			return err
		}

		idem := entities.IdempotencyKey{
			ID:  uuid.New(),
			Key: idempotencyKey,
		}
		if err := tx.Create(&idem).Error; err != nil {
			return err
		}

		return nil
	})
}

func (r *paymentRepositoryImpl) UpdatePaymentWithIdempotency(payment *entities.Payment, idempotencyKey string) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(payment).Error; err != nil {
			return err
		}

		if idempotencyKey != "" {
			idem := entities.IdempotencyKey{
				ID:  uuid.New(),
				Key: idempotencyKey,
			}
			if err := tx.Create(&idem).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *paymentRepositoryImpl) GetPaymentByID(id string) (*entities.Payment, error) {
	var payment entities.Payment
	if err := r.db.Where("id = ?", id).First(&payment).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}
