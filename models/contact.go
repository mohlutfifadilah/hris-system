package models

import (
	"time"

	"github.com/google/uuid"
)

type Contact struct {
	ID                  uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDAddress           *uuid.UUID // id_address
	IDBankAccount       *uuid.UUID // id_address
	No                  string     `gorm:"size:50"`  // no (telepon utama)
	Email               string     `gorm:"size:100"` // email pribadi
	EmergencyConnection string     `gorm:"size:100"` // emergency_connection (nama)
	EmergencyContact    string     `gorm:"size:50"`  // emergency_contact (no hp)
	NoEmergencyContact  string     `gorm:"size:50"`  // no_emergency_contact
	BankAccount         string     `gorm:"size:100"` // bank_account (no rekening utama, kalau mau simple)
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (Contact) TableName() string {
	return "contact"
}
