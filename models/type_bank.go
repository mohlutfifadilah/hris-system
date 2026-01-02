package models

import (
	"time"

	"github.com/google/uuid"
)

type TypeBank struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Type      string    `form:"type" binding:"required" gorm:"size:5;uniqueIndex;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TypeBank) TableName() string {
	return "type_bank"
}
