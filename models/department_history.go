package models

import (
	"time"

	"github.com/google/uuid"
)

type DepartmentHistory struct {
	ID         uuid.UUID `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Department string    `form:"department" binding:"required" gorm:"size:100;not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (DepartmentHistory) TableName() string {
	return "department_history"
}
