package entities

import (
    "time"
    "github.com/google/uuid"
)

type Product struct {
    ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    CategoryID  uuid.UUID `gorm:"type:uuid;not null;index"                       json:"category_id"`
    Name        string    `gorm:"type:varchar(255);not null"                     json:"name"`
    Description string    `gorm:"type:text"                                      json:"description"`
    Price       float64   `gorm:"type:numeric(12,2);not null"                    json:"price"`
	ImageURL    string    `gorm:"type:text"                                      json:"image_url"`
    CreatedAt   time.Time `gorm:"autoCreateTime"                                 json:"created_at"`
    UpdatedAt   time.Time `gorm:"autoUpdateTime"                                 json:"updated_at"`

    // Relationships
    Category Category `gorm:"foreignKey:CategoryID;references:ID" json:"category,omitempty"`
}