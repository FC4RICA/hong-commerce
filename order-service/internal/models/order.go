package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Order struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	UserID             uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	Status             string         `gorm:"type:varchar;not null;default:'PENDING'" json:"status"`
	TotalAmount        float64        `gorm:"type:numeric(12,2);not null" json:"total_amount"`
	PaymentConfirmed   bool           `gorm:"not null;default:false" json:"payment_confirmed"`
	InventoryConfirmed bool           `gorm:"not null;default:false" json:"inventory_confirmed"`
	PaymentID          *uuid.UUID     `gorm:"type:uuid;index" json:"payment_id,omitempty"`
	CreatedAt          time.Time      `gorm:"type:timestamptz;not null" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"type:timestamptz;not null" json:"updated_at"`
	ConfirmedAt        *time.Time     `gorm:"type:timestamptz" json:"confirmed_at,omitempty"`
	Items              []OrderItem    `gorm:"foreignKey:OrderID" json:"items"`
}

func (o *Order) BeforeCreate(tx *gorm.DB) (err error) {
	if o.ID == uuid.Nil {
		o.ID = uuid.New()
	}
	return
}

func (o *Order) BeforeUpdate(tx *gorm.DB) (err error) {
	if o.PaymentConfirmed && o.InventoryConfirmed {
		if o.ConfirmedAt == nil {
			now := time.Now()
			o.ConfirmedAt = &now
		}
		o.Status = "CONFIRMED"
	}
	return
}

type OrderItem struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	OrderID     uuid.UUID `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID   uuid.UUID `gorm:"type:uuid;not null" json:"product_id"`
	ProductName string    `gorm:"type:varchar;not null" json:"product_name"`
	UnitPrice   float64   `gorm:"type:numeric(12,2);not null" json:"unit_price"`
	Quantity    int       `gorm:"type:integer;not null" json:"quantity"`
	Subtotal    float64   `gorm:"type:numeric(12,2);not null" json:"subtotal"`
}

func (oi *OrderItem) BeforeCreate(tx *gorm.DB) (err error) {
	if oi.ID == uuid.Nil {
		oi.ID = uuid.New()
	}
	return
}
