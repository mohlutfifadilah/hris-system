package models

import (
	"time"

	"github.com/google/uuid"
)

type StatusHistory struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Status    string    `gorm:"size:50;not null"` // status
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (StatusHistory) TableName() string {
	return "status_history"
}
