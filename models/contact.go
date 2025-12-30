package models

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID                   uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDAddress            *uuid.UUID // id_address
	NoHp                 string     `form:"no_hp" binding:"required" gorm:"size:100"`                  // no (telepon utama)
	EmergencyRelation    string     `form:"emergency_relation" binding:"required" gorm:"size:100"`     // emergency_relation (hubungan)
	EmergencyContactName string     `form:"emergency_contact_name" binding:"required" gorm:"size:100"` // emergency_contact_name (nama)
	NoEmergencyContact   string     `form:"no_emergency_contact" binding:"required" gorm:"size:100"`   // no_emergency_contact
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

func (Contact) TableName() string {
	return "contact"
}
