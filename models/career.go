package models

import (
	"time"

	"github.com/google/uuid"
)

type Career struct {
	ID                  uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDCareerHistory     *uuid.UUID // id_career_history
	IDCareerAchievement *uuid.UUID // id_career_achievement
	StartWork           string     `gorm:"size:20"` // start_work (bisa string/tanggal)
	PeriodWork          string     `gorm:"size:50"` // period_work
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (Career) TableName() string {
	return "career"
}
