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
	Name                  string               `json:"name"`
	Slug                  string               `json:"slug"`
	Type                  models.ProjectType   `json:"type"`
	BaseURL               string               `json:"base_url"`
	RepoURL               string               `json:"repo_url"`
	SharedValue           int                  `json:"shared_value"`
	DedicatedValue        int                  `json:"dedicated_value"`
	SupportSharedValue    int                  `json:"support_shared_value"`
	SupportDedicatedValue int                  `json:"support_dedicated_value"`
	Status                models.ProjectStatus `json:"status"`
}

type ProjectResponse struct {
	ID                    uuid.UUID            `json:"id"`
	Name                  string               `json:"name"`
	Slug                  string               `json:"slug"`
	Type                  models.ProjectType   `json:"type"`
	BaseURL               string               `json:"base_url"`
	RepoURL               string               `json:"repo_url"`
	SharedValue           int                  `json:"shared_value"`
	DedicatedValue        int                  `json:"dedicated_value"`
	SupportSharedValue    int                  `json:"support_shared_value"`
	SupportDedicatedValue int                  `json:"support_dedicated_value"`
	Status                models.ProjectStatus `json:"status"`
	CreatedAt             time.Time            `json:"created_at"`
	UpdatedAt             time.Time            `json:"updated_at"`
}

func (request ProjectRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if utils.IsBlank(request.Name) || len(strings.TrimSpace(request.Name)) > 100 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_NAME_INVALID"))
	}

	if utils.IsBlank(request.Slug) || len(strings.TrimSpace(request.Slug)) > 100 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_SLUG_INVALID"))
	}

	if !request.Type.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_TYPE_INVALID"))
	}

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("PROJECT_STATUS_INVALID"))
	}

	if len(strings.TrimSpace(request.BaseURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_BASE_URL_INVALID"))
	}

	if len(strings.TrimSpace(request.RepoURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_REPO_URL_INVALID"))
	}

	if request.SharedValue < 0 ||
		request.DedicatedValue < 0 ||
		request.SupportSharedValue < 0 ||
		request.SupportDedicatedValue < 0 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_VALUE_INVALID"))
	}

	return errs
}

func ProjectResponseFromModel(project models.Project) ProjectResponse {
	return ProjectResponse{
		ID:                    project.ID,
		Name:                  project.Name,
		Slug:                  project.Slug,
		Type:                  project.Type,
		BaseURL:               project.BaseURL,
		RepoURL:               project.RepoURL,
		SharedValue:           project.SharedValue,
		DedicatedValue:        project.DedicatedValue,
		SupportSharedValue:    project.SupportSharedValue,
		SupportDedicatedValue: project.SupportDedicatedValue,
		Status:                project.Status,
		CreatedAt:             project.CreatedAt,
		UpdatedAt:             project.UpdatedAt,
	}
}
