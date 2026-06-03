package entities

import (
    "time"
    "github.com/google/uuid"
)

type Category struct {
    ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    Name      string    `gorm:"type:varchar(255);uniqueIndex;not null"         json:"name"`
    CreatedAt time.Time `gorm:"autoCreateTime"                                 json:"created_at"`

    // Relationships
    Products []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}