package migrations

import (
	"hris-system/config"
	"hris-system/models"
)

// RunMigrations menjalankan AutoMigrate untuk semua tabel
func RunMigrations() error {
	db := config.DB

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		&models.Employee{},
		&models.Religion{},
		&models.Blood{},
		&models.BankAccount{},
		&models.Address{},
		&models.Contact{},
		&models.Identity{},
		&models.Education{},
		&models.Staffing{},
		&models.StatusHistory{},
		&models.RankHistory{},
		&models.DepartmentHistory{},
		&models.CareerHistory{},
		&models.TypeAchievement{},
		&models.CareerAchievement{},
		&models.Career{},
	)

}
