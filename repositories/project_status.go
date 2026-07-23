package repositories

import (
	"context"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectStatus interface {
	HasCustomerProject(ctx context.Context, projectID uuid.UUID) (bool, error)
	ListProjectServices(ctx context.Context, projectID uuid.UUID) ([]models.ProjectService, error)
	UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status models.ProjectStatus) error
}

type projectStatus struct {
	db *gorm.DB
}

func NewProjectStatusRepository(db *gorm.DB) ProjectStatus {
	return projectStatus{db: db}
}

func (p projectStatus) HasCustomerProject(ctx context.Context, projectID uuid.UUID) (bool, error) {
	var total int64
	if err := p.db.WithContext(ctx).
		Model(&models.CustomerProject{}).
		Where("project_id = ?", projectID).
		Count(&total).
		Error; err != nil {
		return false, err
	}

	return total > 0, nil
}

func (p projectStatus) ListProjectServices(ctx context.Context, projectID uuid.UUID) ([]models.ProjectService, error) {
	var services []models.ProjectService
	if err := p.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&services).
		Error; err != nil {
		return nil, err
	}

	return services, nil
}

func (p projectStatus) UpdateProjectStatus(ctx context.Context, projectID uuid.UUID, status models.ProjectStatus) error {
	return p.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", projectID).
		Update("status", status).
		Error
}
