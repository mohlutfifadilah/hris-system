// database/seeders/seed.go
package seeders

import (
	"hris-system/config"
	"hris-system/models"
	"hris-system/utils"
	"time"

	"gorm.io/gorm"
)

func seedReligions(tx *gorm.DB) error {
	var count int64
	tx.Model(&models.Religion{}).Count(&count)
	if count > 0 {
		return nil // sudah ada data, skip
	}

	religions := []models.Religion{
		{Religion: "Islam"},
		{Religion: "Kristen"},
		{Religion: "Katolik"},
		{Religion: "Hindu"},
		{Religion: "Buddha"},
		{Religion: "Konghucu"},
	}

	return tx.Create(&religions).Error
}

func seedBloods(tx *gorm.DB) error {
	var count int64
	tx.Model(&models.Blood{}).Count(&count)
	if count > 0 {
		return nil // sudah ada data, skip
	}

	bloods := []models.Blood{
		{BloodType: "A"},
		{BloodType: "B"},
		{BloodType: "AB"},
		{BloodType: "O"},
	}

	return tx.Create(&bloods).Error
}

func seedAdminEmployee(tx *gorm.DB) error {
	// cek apakah admin sudah ada
	var count int64
	tx.Model(&models.Employee{}).
		Where("work_email = ?", "dataanalyse2@plusadvisor.co.id").
		Count(&count)
	if count > 0 {
		return nil
	}

	// ambil 1 religion dan 1 blood (optional, boleh nil)
	var religion models.Religion
	tx.First(&religion)

	var blood models.Blood
	tx.First(&blood)

	// hash password
	passwordHash, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}

	emp := models.Employee{
		WorkEmail:    "dataanalyse2@plusadvisor.co.id",
		Password:     passwordHash,
		Name:         "Moh Lutfi Fadilah",
		Photo:        "null",
		IDEmployee:   "2025.06.1.020",
		Gender:       "Laki-laki",
		Citizenship:  "Indonesia",
		PlaceOfBirth: "Ciamis",
		DateOfBirth:  time.Date(2003, 6, 17, 0, 0, 0, 0, time.UTC),
		Married:      false,
		IDReligion:   &religion.ID,
		IDBlood:      &blood.ID,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return tx.Create(&emp).Error
}

// Seed menjalankan semua seeder (idempotent)
func Seed() error {
	db := config.DB

	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedAdminEmployee(tx); err != nil {
			return err
		}
		if err := seedReligions(tx); err != nil {
			return err
		}
		if err := seedBloods(tx); err != nil {
			return err
		}
		return nil
	})
}
