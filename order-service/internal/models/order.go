package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID         uint           `gorm:"primaryKey" json:"id"`
	UserID     string         `gorm:"not null;index" json:"user_id"`
	ItemID     string         `gorm:"not null" json:"item_id"`
	Quantity   int            `gorm:"not null" json:"quantity"`
	TotalPrice float64        `gorm:"not null" json:"total_price"`
	Status     string         `gorm:"not null;default:'PENDING'" json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
}
