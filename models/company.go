package models

import (
	"time"

	"github.com/google/uuid"
)

type Company struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name          string    `form:"name" binding:"required" gorm:"size:255"`            // name
	NoNpwp        string    `form:"no_npwp" binding:"required" gorm:"size:255"`         // no_npwp
	TaxPersonName string    `form:"tax_person_name" binding:"required" gorm:"size:255"` // tax_person_name
	TaxPersonNpwp string    `form:"tax_person_npwp" binding:"required" gorm:"size:255"` // tax_person_npwp
	Logo          string    `form:"logo" binding:"required" gorm:"size:255"`            // logo
	Address       string    `form:"address" binding:"required" gorm:"size:255"`         // address
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Company) TableName() string {
	return "company"
}
