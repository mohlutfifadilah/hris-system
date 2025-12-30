package models

import (
	"time"

	"github.com/google/uuid"
)

type Career struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDEmployee    *uuid.UUID // id_employee
	IDStatus      *uuid.UUID // id_status
	IDGrading     *uuid.UUID // id_grading
	IDDepartment  *uuid.UUID // id_department
	EffectiveDate time.Time  // effective_date
	Position      string     `form:"position" binding:"required" gorm:"size:255"`      // position
	EvidenceLink  string     `form:"evidence_link" binding:"required" gorm:"size:255"` // evidence_link (file path / url)
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Career) TableName() string {
	return "career_history"
}
