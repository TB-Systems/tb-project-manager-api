package dto

import (
	"strings"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

type ProjectServiceListFilter struct {
	ProjectID string
}

type ProjectServiceCreateRequest struct {
	ProjectID      uuid.UUID          `json:"project_id"`
	Name           string             `json:"name"`
	Type           models.ProjectType `json:"type"`
	URL            string             `json:"url"`
	RepoURL        string             `json:"repo_url"`
	HealthCheckURL string             `json:"health_check_url"`
}

type ProjectServiceRequest struct {
	ProjectID      uuid.UUID            `json:"project_id"`
	Name           string               `json:"name"`
	Type           models.ProjectType   `json:"type"`
	URL            string               `json:"url"`
	RepoURL        string               `json:"repo_url"`
	Status         models.ProjectStatus `json:"status"`
	HealthCheckURL string               `json:"health_check_url"`
}

type ProjectServiceResponse struct {
	ID             uuid.UUID            `json:"id"`
	ProjectID      uuid.UUID            `json:"project_id"`
	Name           string               `json:"name"`
	Type           models.ProjectType   `json:"type"`
	URL            string               `json:"url"`
	RepoURL        string               `json:"repo_url"`
	Status         models.ProjectStatus `json:"status"`
	HealthCheckURL string               `json:"health_check_url"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

func (request ProjectServiceRequest) Validate() []errors.ApiErrorItem {
	errs := validateProjectServiceFields(projectServiceValidationFields{
		ProjectID:      request.ProjectID,
		Name:           request.Name,
		Type:           request.Type,
		URL:            request.URL,
		RepoURL:        request.RepoURL,
		HealthCheckURL: request.HealthCheckURL,
	})

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_STATUS_INVALID"))
	}

	return errs
}

func (request ProjectServiceCreateRequest) Validate() []errors.ApiErrorItem {
	return validateProjectServiceFields(projectServiceValidationFields{
		ProjectID:      request.ProjectID,
		Name:           request.Name,
		Type:           request.Type,
		URL:            request.URL,
		RepoURL:        request.RepoURL,
		HealthCheckURL: request.HealthCheckURL,
	})
}

type projectServiceValidationFields struct {
	ProjectID      uuid.UUID
	Name           string
	Type           models.ProjectType
	URL            string
	RepoURL        string
	HealthCheckURL string
}

func validateProjectServiceFields(fields projectServiceValidationFields) []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if fields.ProjectID == uuid.Nil {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_PROJECT_ID_INVALID"))
	}

	if utils.IsBlank(fields.Name) || len(strings.TrimSpace(fields.Name)) > 100 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_NAME_INVALID"))
	}

	if !fields.Type.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_TYPE_INVALID"))
	}

	if len(strings.TrimSpace(fields.URL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_URL_INVALID"))
	}

	if len(strings.TrimSpace(fields.RepoURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_REPO_URL_INVALID"))
	}

	if len(strings.TrimSpace(fields.HealthCheckURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_HEALTH_CHECK_URL_INVALID"))
	}

	return errs
}

func ProjectServiceResponseFromModel(projectService models.ProjectService) ProjectServiceResponse {
	return ProjectServiceResponse{
		ID:             projectService.ID,
		ProjectID:      projectService.ProjectID,
		Name:           projectService.Name,
		Type:           projectService.Type,
		URL:            projectService.URL,
		RepoURL:        projectService.RepoURL,
		Status:         projectService.Status,
		HealthCheckURL: projectService.HealthCheckURL,
		CreatedAt:      projectService.CreatedAt,
		UpdatedAt:      projectService.UpdatedAt,
	}
}
