package models

import (
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID            uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDCompany     *uuid.UUID
	IDStaffing    *uuid.UUID
	IDContact     *uuid.UUID
	IDIdentity    *uuid.UUID
	IDBankAccount *uuid.UUID
	IDBlood       *uuid.UUID
	IDFamily      *uuid.UUID
	IDReligion    *uuid.UUID
	WorkEmail     string `form:"work_email" binding:"required" gorm:"size:100;uniqueIndex;not null"` // ← email kantor
	Email         string `form:"email" binding:"required" gorm:"size:100;uniqueIndex;not null"`      // ← email pribadi
	Password      string `form:"password" binding:"required" gorm:"size:255;not null"`
	Name          string `form:"name" binding:"required" gorm:"size:150;not null"`
	Photo         string `form:"photo" binding:"required" gorm:"size:255"`
	IDEmployee    string `form:"id_employee" binding:"required" gorm:"size:50;uniqueIndex"`
	Gender        string `form:"gender" binding:"required" gorm:"size:10"`
	Citizenship   string `form:"citizenship" binding:"required" gorm:"size:50"`
	PlaceOfBirth  string `form:"place_of_birth" binding:"required" gorm:"size:100"`
	DateOfBirth   time.Time
	MarialStatus  string `form:"marial_status" binding:"required" gorm:"size:10"`
	JoinDate      string `form:"join_date" binding:"required" gorm:"size:10"`
	EvidenceLink  string `form:"evidence_link" binding:"required" gorm:"size:10"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Employee) TableName() string {
	return "employee"
}
