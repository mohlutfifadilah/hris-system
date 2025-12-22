package models

import (
	"time"

	"github.com/google/uuid"
)

type CareerAchievement struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDType      *uuid.UUID // id_type
	Date        time.Time  // date
	Title       string     `gorm:"size:150"` // title
	Description string     `gorm:"size:255"` // description
	Evidence    string     `gorm:"size:255"` // evidence (file path / url)
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (CareerAchievement) TableName() string {
	return "career_achievement"
}
