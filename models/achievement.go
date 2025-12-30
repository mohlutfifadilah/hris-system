package models

import (
	"time"

	"github.com/google/uuid"
)

type Achievement struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDEmployee   *uuid.UUID // id_employee
	IDType       *uuid.UUID // id_type
	Date         time.Time  `form:"date" binding:"required"`                          // date
	Title        string     `form:"title" binding:"required" gorm:"size:150"`         // title
	Description  string     `form:"description" binding:"required" gorm:"size:255"`   // description
	EvidenceLink string     `form:"evidence_link" binding:"required" gorm:"size:255"` // evidence_link (file path / url)
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Achievement) TableName() string {
	return "achievement"
}
