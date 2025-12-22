package models

import (
	"time"

	"github.com/google/uuid"
)

type Address struct {
	ID              uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	AddressIdentity string    `gorm:"size:255"` // address_identity
	AddressDomicile string    `gorm:"size:255"` // address_domicile
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func (Address) TableName() string {
	return "address"
}
