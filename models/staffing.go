package models

import (
	"time"

	"github.com/google/uuid"
)

type Staffing struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	NoBpjsKes       string    `form:"no_bpjs_kes" binding:"required" gorm:"size:50"`       // no_bpjs
	NoBpjsJk        string    `form:"no_bpjs_jk" binding:"required" gorm:"size:50"`        // no_bpjs
	NoKjpBpjs       string    `form:"no_kjp_bpjs" binding:"required" gorm:"size:50"`       // kjp_bpjs
	NoNpwpLimabelas string    `form:"no_npwp_limabelas" binding:"required" gorm:"size:50"` // no_npwp
	NoNpwpEnambelas string    `form:"no_npwp_enambelas" binding:"required" gorm:"size:50"` // no_npwp
	Ptkp            string    `form:"ptkp" binding:"required" gorm:"size:50"`              // tax_payer
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Staffing) TableName() string {
	return "staffing"
}
