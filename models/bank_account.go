package models

import (
	"time"

	"github.com/google/uuid"
)

type BankAccount struct {
	ID        uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Bank      string    `gorm:"size:100;not null"` // bank
	NoAccount string    `gorm:"size:100;not null"` // no_account
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (BankAccount) TableName() string {
	return "bank_account"
}
