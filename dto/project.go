package dto

import (
	"strings"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

type ProjectRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Slug        string               `json:"slug"`
	RepoURL     string               `json:"repo_url"`
	Status      models.ProjectStatus `json:"status"`
}

type ProjectResponse struct {
	ID          uuid.UUID            `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Slug        string               `json:"slug"`
	RepoURL     string               `json:"repo_url"`
	Status      models.ProjectStatus `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

func (request ProjectRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if utils.IsBlank(request.Name) || len(strings.TrimSpace(request.Name)) > 100 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_NAME_INVALID"))
	}

	if utils.IsBlank(request.Slug) || len(strings.TrimSpace(request.Slug)) > 50 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SLUG_INVALID"))
	}

	if len(strings.TrimSpace(request.Description)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_DESCRIPTION_INVALID"))
	}

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_STATUS_INVALID"))
	}

	if len(strings.TrimSpace(request.RepoURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_REPO_URL_INVALID"))
	}

	return errs
}

func ProjectResponseFromModel(project models.Project) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Slug:        project.Slug,
		RepoURL:     project.RepoURL,
		Status:      project.Status,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}
