package models

import (
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Type      string    `gorm:"size:50;not null"`  // type (KTP, KITAS, dll)
	No        string    `gorm:"size:100;not null"` // no
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Identity) TableName() string {
	return "identity"
}
