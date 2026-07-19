package dto

import (
	"strings"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

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
	errs := make([]errors.ApiErrorItem, 0)

	if request.ProjectID == uuid.Nil {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_PROJECT_ID_INVALID"))
	}

	if utils.IsBlank(request.Name) || len(strings.TrimSpace(request.Name)) > 100 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_NAME_INVALID"))
	}

	if !request.Type.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_TYPE_INVALID"))
	}

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_STATUS_INVALID"))
	}

	if len(strings.TrimSpace(request.URL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_URL_INVALID"))
	}

	if len(strings.TrimSpace(request.RepoURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SERVICE_REPO_URL_INVALID"))
	}

	if len(strings.TrimSpace(request.HealthCheckURL)) > 500 {
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
