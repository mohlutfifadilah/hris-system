// database/seeders/seed.go
package seeders

import (
	"hris-system/config"
	"hris-system/models"
	"hris-system/utils"
	"time"

	"gorm.io/gorm"
)

func seedCompany(tx *gorm.DB) error {

	emp := models.Company{
		Name:          "PT. Trinitas Strategis Solusi",
		NoNpwp:        "516376123456789",
		TaxPersonName: "Admin",
		TaxPersonNpwp: "Admin",
		Logo:          "/static/img/logo.png",
		Address:       "Jakarta Selatan",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	return tx.Create(&emp).Error
}

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

func seedTypeBanks(tx *gorm.DB) error {
	var count int64
	tx.Model(&models.TypeBank{}).Count(&count)
	if count > 0 {
		return nil // sudah ada data, skip
	}

	typeBanks := []models.TypeBank{
		{Type: "BCA"},
		{Type: "BNI"},
		{Type: "BRI"},
		{Type: "Mandiri"},
	}

	return tx.Create(&typeBanks).Error
}

func seedTypeIdentity(tx *gorm.DB) error {
	var count int64
	tx.Model(&models.TypeIdentity{}).Count(&count)
	if count > 0 {
		return nil // sudah ada data, skip
	}

	typeIdentities := []models.TypeIdentity{
		{Type: "KTP"},
		{Type: "SIM"},
		{Type: "Passport"},
	}

	return tx.Create(&typeIdentities).Error
}

func seedAdminEmployee(tx *gorm.DB) error {
	// cek apakah admin sudah ada
	var count int64
	tx.Model(&models.Employee{}).
		Where("work_email = ?", "admin@plusadvisor.co.id").
		Count(&count)
	if count > 0 {
		return nil
	}

	// ambil 1 religion dan 1 blood (optional, boleh nil)
	var religion models.Religion
	tx.First(&religion)

	var blood models.Blood
	tx.First(&blood)

	var company models.Company
	tx.First(&company)

	// hash password
	passwordHash, err := utils.HashPassword("admin123")
	if err != nil {
		return err
	}

	emp := models.Employee{
		IDCompany:    &company.ID,
		IDBlood:      &blood.ID,
		IDReligion:   &religion.ID,
		WorkEmail:    "admin@plusadvisor.co.id",
		Password:     passwordHash,
		Name:         "Admin",
		Photo:        "/static/img/logo.png",
		IDEmployee:   "0000.00.0.000",
		Gender:       "Male",
		Citizenship:  "Indonesia",
		PlaceOfBirth: "Ciamis",
		DateOfBirth:  time.Date(2003, 6, 17, 0, 0, 0, 0, time.UTC),
		MaritalStatus: false,
		IsAdmin: true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	return tx.Create(&emp).Error
}

// Seed menjalankan semua seeder (idempotent)
func Seed() error {
	db := config.DB

	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedReligions(tx); err != nil {
			return err
		}
		if err := seedBloods(tx); err != nil {
			return err
		}
		if err := seedCompany(tx); err != nil {
			return err
		}
		if err := seedTypeBanks(tx); err != nil {
			return err
		}
		if err := seedTypeIdentity(tx); err != nil {
			return err
		}
		if err := seedAdminEmployee(tx); err != nil {
			return err
		}
		return nil
	})
}
