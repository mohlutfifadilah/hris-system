package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID          uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDEducation *uuid.UUID
	IDCareer    *uuid.UUID
	IDStaffing  *uuid.UUID
	IDContact   *uuid.UUID
	IDIdentity  *uuid.UUID
	IDReligion  *uuid.UUID
	IDBlood     *uuid.UUID

	WorkEmail string `gorm:"size:100;uniqueIndex;not null"` // ← untuk login
	Password  string `gorm:"size:255;not null"`

	Name         string `gorm:"size:150;not null"`
	Photo        string `gorm:"size:255"`
	IDEmployee   string `gorm:"size:50;uniqueIndex"`
	Gender       string `gorm:"size:10"`
	Citizenship  string `gorm:"size:50"`
	PlaceOfBirth string `gorm:"size:100"`
	DateOfBirth  time.Time
	Married      bool

	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Employee) TableName() string {
	return "employee"
}
