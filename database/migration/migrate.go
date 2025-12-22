package migrations

import (
	"hris-system/config"
	"hris-system/models"
)

// RunMigrations menjalankan AutoMigrate untuk semua tabel
func RunMigrations() error {
	db := config.DB

	// enable extension
	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return err
	}

	return db.AutoMigrate(
		// reference/master
		&models.Religion{},
		&models.Blood{},
		&models.BankAccount{},
		&models.Address{},
		&models.Contact{},
		&models.Identity{},
		&models.Education{},
		&models.Staffing{},

		// career & history
		&models.StatusHistory{},
		&models.RankHistory{},
		&models.DepartmentHistory{},
		&models.CareerHistory{},
		&models.TypeAchievement{},
		&models.CareerAchievement{},
		&models.Career{},

		// main
		&models.Employee{},
	)
}
