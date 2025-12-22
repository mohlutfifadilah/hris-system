package models

import (
	"time"

	"github.com/google/uuid"
)

type Education struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	LastEducation string    `gorm:"size:100"` // last_education
	Major         string    `gorm:"size:100"` // major
	InstituteName string    `gorm:"size:150"` // institute_name
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Education) TableName() string {
	return "education"
}
