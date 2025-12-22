package models

import (
	"time"

	"github.com/google/uuid"
)

type Staffing struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	NoBpjs     string    `gorm:"size:50"` // no_bpjs
	KjpBpjs    string    `gorm:"size:50"` // kjp_bpjs
	NoNpwp     string    `gorm:"size:50"` // no_npwp
	TaxPayer   string    `gorm:"size:50"` // tax_payer
	CutterNpwp string    `gorm:"size:50"` // cutter_npwp
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (Staffing) TableName() string {
	return "staffing"
}
