package repositories

import (
	"context"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerProject interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.CustomerProject, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.CustomerProject, error)
	Create(ctx context.Context, customerProject models.CustomerProject) (models.CustomerProject, error)
	Update(ctx context.Context, customerProject models.CustomerProject) (models.CustomerProject, error)
	Delete(ctx context.Context, id uuid.UUID) error
	LinkExists(ctx context.Context, projectID uuid.UUID, customerID uuid.UUID, exceptID *uuid.UUID) (bool, error)
}

type customerProject struct {
	db *gorm.DB
}

func NewCustomerProjectRepository(db *gorm.DB) CustomerProject {
	return customerProject{db: db}
}

func (c customerProject) List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.CustomerProject, int64, error) {
	var customerProjects []models.CustomerProject
	var total int64

	query := c.db.WithContext(ctx).Model(&models.CustomerProject{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(&customerProjects).
		Error
	if err != nil {
		return nil, 0, err
	}

	return customerProjects, total, nil
}

func (c customerProject) FindByID(ctx context.Context, id uuid.UUID) (models.CustomerProject, error) {
	var customerProject models.CustomerProject
	if err := c.db.WithContext(ctx).First(&customerProject, "id = ?", id).Error; err != nil {
		return models.CustomerProject{}, err
	}

	return customerProject, nil
}

func (c customerProject) Create(ctx context.Context, customerProject models.CustomerProject) (models.CustomerProject, error) {
	if err := c.db.WithContext(ctx).Create(&customerProject).Error; err != nil {
		return models.CustomerProject{}, err
	}

	return customerProject, nil
}

func (c customerProject) Update(ctx context.Context, customerProject models.CustomerProject) (models.CustomerProject, error) {
	err := c.db.WithContext(ctx).
		Model(&models.CustomerProject{}).
		Where("id = ?", customerProject.ID).
		Updates(map[string]interface{}{
			"project_id":             customerProject.ProjectID,
			"customer_id":            customerProject.CustomerID,
			"project_value":          customerProject.ProjectValue,
			"monthly_value":          customerProject.MonthlyValue,
			"due_day":                customerProject.DueDay,
			"project_payment_status": customerProject.ProjectPaymentStatus,
			"last_payment":           customerProject.LastPayment,
		}).
		Error
	if err != nil {
		return models.CustomerProject{}, err
	}

	return c.FindByID(ctx, customerProject.ID)
}

func (c customerProject) Delete(ctx context.Context, id uuid.UUID) error {
	return c.db.WithContext(ctx).Delete(&models.CustomerProject{}, "id = ?", id).Error
}

func (c customerProject) LinkExists(ctx context.Context, projectID uuid.UUID, customerID uuid.UUID, exceptID *uuid.UUID) (bool, error) {
	var total int64

	query := c.db.WithContext(ctx).
		Model(&models.CustomerProject{}).
		Where("project_id = ? AND customer_id = ?", projectID, customerID)
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}

	if err := query.Count(&total).Error; err != nil {
		return false, err
	}

	return total > 0, nil
}
