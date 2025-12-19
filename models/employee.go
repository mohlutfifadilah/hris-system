package models

import (
	"time"

	"gorm.io/gorm"
)

type Employee struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Nama      string         `gorm:"column:nama" json:"nama"`
	Email     string         `gorm:"column:email" json:"email"`
	Unit      string         `gorm:"column:unit" json:"unit"`
	Password  string         `gorm:"column:password" json:"-"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Employee) TableName() string {
	return "employees"
}
