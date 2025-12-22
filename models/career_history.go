package models

import (
	"time"

	"github.com/google/uuid"
)

type CareerHistory struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDStatus      *uuid.UUID // id_status
	IDRank        *uuid.UUID // id_rank
	IDDepartment  *uuid.UUID // id_department
	Placement     string     `gorm:"size:100"` // placement
	EffectiveDate time.Time  // effective_date
	Position      string     `gorm:"size:100"` // position
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (CareerHistory) TableName() string {
	return "career_history"
}
