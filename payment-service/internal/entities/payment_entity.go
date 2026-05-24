package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Payment struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	OrderID        string         `gorm:"uniqueIndex" json:"order_id"`
	UserID         string         `json:"user_id"`
	Amount         float64        `json:"amount"`
	Status         string         `json:"status"`
	Currency       string         `json:"currency"`
	TransactionRef string         `json:"transaction_ref"`
}

// BeforeCreate is a GORM hook that automatically runs before inserting a record to populate the UUID if it is not already set.
func (p *Payment) BeforeCreate(tx *gorm.DB) (err error) {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return
}
