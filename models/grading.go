package models

import (
	"time"

	"github.com/google/uuid"
)

type Grading struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Grading   string    `form:"grading" binding:"required" gorm:"size:50;not null"` // grading
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Grading) TableName() string {
	return "grading"
}
