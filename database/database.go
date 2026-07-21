package database

import (
	"github.com/TB-Systems/tb-project-manager-api/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	if err := dropLegacyProjectColumns(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.Customer{},
		&models.Project{},
		&models.CustomerProject{},
		&models.ProjectService{},
		&models.ServiceCheck{},
		&models.ServiceLog{},
	)
}

func dropLegacyProjectColumns(db *gorm.DB) error {
	columns := []string{
		"type",
		"base_url",
		"shared_value",
		"dedicated_value",
		"support_shared_value",
		"support_dedicated_value",
	}

	migrator := db.Migrator()
	for _, column := range columns {
		if migrator.HasColumn(&models.Project{}, column) {
			if err := migrator.DropColumn(&models.Project{}, column); err != nil {
				return err
			}
		}
	}

	return nil
}
