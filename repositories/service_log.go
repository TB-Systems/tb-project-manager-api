package repositories

import (
	"context"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceLog interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.ServiceLog, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.ServiceLog, error)
	Create(ctx context.Context, serviceLog models.ServiceLog) (models.ServiceLog, error)
}

type serviceLog struct {
	db *gorm.DB
}

func NewServiceLogRepository(db *gorm.DB) ServiceLog {
	return serviceLog{db: db}
}

func (s serviceLog) List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.ServiceLog, int64, error) {
	var serviceLogs []models.ServiceLog
	var total int64

	query := s.db.WithContext(ctx).Model(&models.ServiceLog{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("time DESC").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(&serviceLogs).
		Error
	if err != nil {
		return nil, 0, err
	}

	return serviceLogs, total, nil
}

func (s serviceLog) FindByID(ctx context.Context, id uuid.UUID) (models.ServiceLog, error) {
	var serviceLog models.ServiceLog
	if err := s.db.WithContext(ctx).First(&serviceLog, "id = ?", id).Error; err != nil {
		return models.ServiceLog{}, err
	}

	return serviceLog, nil
}

func (s serviceLog) Create(ctx context.Context, serviceLog models.ServiceLog) (models.ServiceLog, error) {
	if err := s.db.WithContext(ctx).Create(&serviceLog).Error; err != nil {
		return models.ServiceLog{}, err
	}

	return serviceLog, nil
}
