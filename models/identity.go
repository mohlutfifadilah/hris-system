package models

import (
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	ID           uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Type         string    `form:"type" binding:"required" gorm:"size:50;not null"`           // type (KTP, KITAS, dll)
	No           string    `form:"no" binding:"required" gorm:"size:100;not null"`            // no
	EvidenceLink string    `form:"evidence_link" binding:"required" gorm:"size:100;not null"` // no
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Identity) TableName() string {
	return "identity"
}
