package repositories

import (
	"context"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceCheck interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.ServiceCheck, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.ServiceCheck, error)
	Create(ctx context.Context, serviceCheck models.ServiceCheck) (models.ServiceCheck, error)
	Update(ctx context.Context, serviceCheck models.ServiceCheck) (models.ServiceCheck, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type serviceCheck struct {
	db *gorm.DB
}

func NewServiceCheckRepository(db *gorm.DB) ServiceCheck {
	return serviceCheck{db: db}
}

func (s serviceCheck) List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.ServiceCheck, int64, error) {
	var serviceChecks []models.ServiceCheck
	var total int64

	query := s.db.WithContext(ctx).Model(&models.ServiceCheck{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(&serviceChecks).
		Error
	if err != nil {
		return nil, 0, err
	}

	return serviceChecks, total, nil
}

func (s serviceCheck) FindByID(ctx context.Context, id uuid.UUID) (models.ServiceCheck, error) {
	var serviceCheck models.ServiceCheck
	if err := s.db.WithContext(ctx).First(&serviceCheck, "id = ?", id).Error; err != nil {
		return models.ServiceCheck{}, err
	}

	return serviceCheck, nil
}

func (s serviceCheck) Create(ctx context.Context, serviceCheck models.ServiceCheck) (models.ServiceCheck, error) {
	if err := s.db.WithContext(ctx).Create(&serviceCheck).Error; err != nil {
		return models.ServiceCheck{}, err
	}

	return serviceCheck, nil
}

func (s serviceCheck) Update(ctx context.Context, serviceCheck models.ServiceCheck) (models.ServiceCheck, error) {
	err := s.db.WithContext(ctx).
		Model(&models.ServiceCheck{}).
		Where("id = ?", serviceCheck.ID).
		Updates(map[string]interface{}{
			"project_service_id": serviceCheck.ProjectServiceID,
			"status":             serviceCheck.Status,
			"status_code":        serviceCheck.StatusCode,
			"response_time_ms":   serviceCheck.ResponseTimeMS,
			"message":            serviceCheck.Message,
		}).
		Error
	if err != nil {
		return models.ServiceCheck{}, err
	}

	return s.FindByID(ctx, serviceCheck.ID)
}

func (s serviceCheck) Delete(ctx context.Context, id uuid.UUID) error {
	return s.db.WithContext(ctx).Delete(&models.ServiceCheck{}, "id = ?", id).Error
}
