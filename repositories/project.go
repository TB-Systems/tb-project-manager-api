package repositories

import (
	"context"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Project interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.Project, int64, error)
	Overview(ctx context.Context) ([]models.Project, error)
	FindByID(ctx context.Context, id uuid.UUID) (models.Project, error)
	Create(ctx context.Context, project models.Project) (models.Project, error)
	Update(ctx context.Context, project models.Project) (models.Project, error)
	Delete(ctx context.Context, id uuid.UUID) error
	SlugExists(ctx context.Context, slug string, exceptID *uuid.UUID) (bool, error)
}

type project struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) Project {
	return project{db: db}
}

func (p project) List(ctx context.Context, params commonsmodels.PaginatedParams) ([]models.Project, int64, error) {
	var projects []models.Project
	var total int64

	query := p.db.WithContext(ctx).Model(&models.Project{})
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("created_at DESC").
		Limit(int(params.Limit)).
		Offset(int(params.Offset)).
		Find(&projects).
		Error
	if err != nil {
		return nil, 0, err
	}

	return projects, total, nil
}

func (p project) Overview(ctx context.Context) ([]models.Project, error) {
	var projects []models.Project

	query := p.db.WithContext(ctx).Model(&models.Project{})
	err := query.
		Preload("CustomerProjects", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("CustomerProjects.Customer").
		Preload("Services", func(db *gorm.DB) *gorm.DB {
			return db.Order("name ASC")
		}).
		Order("updated_at DESC").
		Find(&projects).
		Error
	if err != nil {
		return nil, err
	}

	return projects, nil
}

func (p project) FindByID(ctx context.Context, id uuid.UUID) (models.Project, error) {
	var project models.Project
	if err := p.db.WithContext(ctx).
		Preload("CustomerProjects", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("CustomerProjects.Customer").
		First(&project, "id = ?", id).Error; err != nil {
		return models.Project{}, err
	}

	return project, nil
}

func (p project) Create(ctx context.Context, project models.Project) (models.Project, error) {
	if err := p.db.WithContext(ctx).Create(&project).Error; err != nil {
		return models.Project{}, err
	}

	return project, nil
}

func (p project) Update(ctx context.Context, project models.Project) (models.Project, error) {
	err := p.db.WithContext(ctx).
		Model(&models.Project{}).
		Where("id = ?", project.ID).
		Updates(map[string]interface{}{
			"name":        project.Name,
			"description": project.Description,
			"slug":        project.Slug,
			"repo_url":    project.RepoURL,
			"status":      project.Status,
		}).
		Error
	if err != nil {
		return models.Project{}, err
	}

	return p.FindByID(ctx, project.ID)
}

func (p project) Delete(ctx context.Context, id uuid.UUID) error {
	return p.db.WithContext(ctx).Delete(&models.Project{}, "id = ?", id).Error
}

func (p project) SlugExists(ctx context.Context, slug string, exceptID *uuid.UUID) (bool, error) {
	var total int64

	query := p.db.WithContext(ctx).Model(&models.Project{}).Where("slug = ?", slug)
	if exceptID != nil {
		query = query.Where("id <> ?", *exceptID)
	}

	if err := query.Count(&total).Error; err != nil {
		return false, err
	}

	return total > 0, nil
}
