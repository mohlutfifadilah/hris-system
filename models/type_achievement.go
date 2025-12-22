package models

import (
	"time"

	"github.com/google/uuid"
)

type TypeAchievement struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Type      string    `gorm:"size:100;not null"` // type
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (TypeAchievement) TableName() string {
	return "type_achievement"
}
