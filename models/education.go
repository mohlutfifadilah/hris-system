package models

import (
	"time"

	"github.com/google/uuid"
)

type Education struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	IDEmployee    *uuid.UUID // id_employee
	Education     string     `form:"education" binding:"required" gorm:"size:100"`      // last_education
	Major         string     `form:"major" binding:"required" gorm:"size:100"`          // major
	InstituteName string     `form:"institute_name" binding:"required" gorm:"size:150"` // institute_name
	IPK           string     `form:"ipk" binding:"required" gorm:"size:150"`            // institute_name
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Education) TableName() string {
	return "education"
}
