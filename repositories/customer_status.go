package repositories

import (
	"context"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerStatus interface {
	ListCustomerProjects(ctx context.Context, customerID uuid.UUID) ([]models.CustomerProject, error)
	UpdateCustomerStatus(ctx context.Context, customerID uuid.UUID, status models.CustomerStatus) error
}

type customerStatus struct {
	db *gorm.DB
}

func NewCustomerStatusRepository(db *gorm.DB) CustomerStatus {
	return customerStatus{db: db}
}

func (c customerStatus) ListCustomerProjects(ctx context.Context, customerID uuid.UUID) ([]models.CustomerProject, error) {
	var customerProjects []models.CustomerProject
	err := c.db.WithContext(ctx).
		Preload("Invoices").
		Where("customer_id = ?", customerID).
		Find(&customerProjects).
		Error
	if err != nil {
		return nil, err
	}

	return customerProjects, nil
}

func (c customerStatus) UpdateCustomerStatus(ctx context.Context, customerID uuid.UUID, status models.CustomerStatus) error {
	return c.db.WithContext(ctx).
		Model(&models.Customer{}).
		Where("id = ?", customerID).
		Update("status", status).
		Error
}
