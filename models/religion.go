package models

import (
	"time"

	"github.com/google/uuid"
)

type Religion struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Religion  string    `gorm:"size:255;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Religion) TableName() string {
	return "religion"
}
