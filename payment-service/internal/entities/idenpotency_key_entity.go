package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type IdempotencyKey struct {
	ID           uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	Key          string         `json:"key"`
	ResponseCode int            `json:"response_code"`
	ResponseBody string         `json:"response_body"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	ExpiresAt    time.Time      `json:"expires_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate is a GORM hook that automatically runs before inserting a record to populate the UUID if it is not already set.
func (ik *IdempotencyKey) BeforeCreate(tx *gorm.DB) (err error) {
	if ik.ID == uuid.Nil {
		ik.ID = uuid.New()
	}
	return
}
