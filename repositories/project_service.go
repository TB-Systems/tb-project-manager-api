package repositories

import (
	"context"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectService interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.ProjectService, int64, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.ProjectService, error)
	Create(ctx context.Context, projectService models.ProjectService) (models.ProjectService, error)
	Update(ctx context.Context, projectService models.ProjectService) (models.ProjectService, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type projectService struct {
	db *gorm.DB
}

func NewProjectServiceRepository(db *gorm.DB) ProjectService {
	return projectService{db: db}
}

func (p projectService) List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.ProjectService, int64, error) {
	var projectServices []models.ProjectService
	var total int64

	query := p.db.WithContext(ctx).Model(&models.ProjectService{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(&projectServices).
		Error
	if err != nil {
		return nil, 0, err
	}

	return projectServices, total, nil
}

func (p projectService) FindByID(ctx context.Context, id uuid.UUID) (models.ProjectService, error) {
	var projectService models.ProjectService
	if err := p.db.WithContext(ctx).First(&projectService, "id = ?", id).Error; err != nil {
		return models.ProjectService{}, err
	}

	return projectService, nil
}

func (p projectService) Create(ctx context.Context, projectService models.ProjectService) (models.ProjectService, error) {
	if err := p.db.WithContext(ctx).Create(&projectService).Error; err != nil {
		return models.ProjectService{}, err
	}

	return projectService, nil
}

func (p projectService) Update(ctx context.Context, projectService models.ProjectService) (models.ProjectService, error) {
	err := p.db.WithContext(ctx).
		Model(&models.ProjectService{}).
		Where("id = ?", projectService.ID).
		Updates(map[string]interface{}{
			"project_id":       projectService.ProjectID,
			"name":             projectService.Name,
			"type":             projectService.Type,
			"url":              projectService.URL,
			"repo_url":         projectService.RepoURL,
			"status":           projectService.Status,
			"health_check_url": projectService.HealthCheckURL,
		}).
		Error
	if err != nil {
		return models.ProjectService{}, err
	}

	return p.FindByID(ctx, projectService.ID)
}

func (p projectService) Delete(ctx context.Context, id uuid.UUID) error {
	return p.db.WithContext(ctx).Delete(&models.ProjectService{}, "id = ?", id).Error
}
