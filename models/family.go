package models

import (
	"time"

	"github.com/google/uuid"
)

type Family struct {
	ID           uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDEmployee   *uuid.UUID // id_employee
	IDReligion   *uuid.UUID // id_religion
	Name         string     `form:"name" binding:"required" gorm:"size:150"`     // name
	Relation     string     `form:"relation" binding:"required" gorm:"size:255"` // relation
	DateOfBirth  time.Time  `form:"date_of_birth" binding:"required"`            // date_of_birth
	MarialStatus string     `form:"marial_status" binding:"required" gorm:"size:10"`
	Gender       string     `form:"gender" binding:"required" gorm:"size:10"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (Family) TableName() string {
	return "family"
}
