package migrations

import (
	"fmt"
	"hris-system/config"
	"hris-system/models"
)

// RunMigrations menjalankan AutoMigrate untuk semua tabel
func RunMigrations() error {
	db := config.DB

	if err := db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp";`).Error; err != nil {
		return err
	}

	models := []interface{}{
        &models.Company{},
        &models.Religion{},
        &models.Blood{},
        &models.TypeBank{},
        &models.TypeIdentity{},
        &models.Status{},
        &models.Grading{},
        &models.Department{},
        &models.TypeAchievement{},
        &models.BankAccount{},
        &models.Address{},
        &models.Contact{},
        &models.Identity{},
        &models.Education{},
        &models.Staffing{},
        &models.Employee{},
        &models.Achievement{},
        &models.Career{},
        &models.Family{},
    }

    for _, model := range models {
        if err := db.AutoMigrate(model); err != nil {
            return fmt.Errorf("failed to migrate %T: %w", model, err)
        }
    }

    return nil

}
