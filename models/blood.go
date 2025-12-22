package models

import (
	"time"

	"github.com/google/uuid"
)

type Blood struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BloodType string    `gorm:"size:5;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Blood) TableName() string {
	return "blood"
}
