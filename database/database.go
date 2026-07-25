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
	if err := dropLegacyCustomerProjectBillingColumns(db); err != nil {
		return err
	}
	if err := dropLegacyCustomerProjectInvoiceTypeConstraint(db); err != nil {
		return err
	}
	if err := dropLegacyCustomerProjectProjectUniqueIndex(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		&models.User{},
		&models.UserSession{},
		&models.Customer{},
		&models.Project{},
		&models.CustomerProject{},
		&models.CustomerProjectTerm{},
		&models.CustomerProjectInvoice{},
		&models.ProjectService{},
		&models.ServiceCheck{},
		&models.ServiceLog{},
	)
}

func dropLegacyCustomerProjectProjectUniqueIndex(db *gorm.DB) error {
	return db.Exec("DROP INDEX IF EXISTS idx_customer_projects_project_id").Error
}

func dropLegacyCustomerProjectInvoiceTypeConstraint(db *gorm.DB) error {
	return db.Exec("ALTER TABLE customer_project_invoices DROP CONSTRAINT IF EXISTS chk_customer_project_invoices_type").Error
}

func dropLegacyCustomerProjectBillingColumns(db *gorm.DB) error {
	columns := []string{
		"project_value",
		"monthly_value",
		"due_day",
		"project_payment_status",
		"last_payment",
	}

	migrator := db.Migrator()
	for _, column := range columns {
		if migrator.HasColumn(&models.CustomerProject{}, column) {
			if err := migrator.DropColumn(&models.CustomerProject{}, column); err != nil {
				return err
			}
		}
	}

	return nil
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
