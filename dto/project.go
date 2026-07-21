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

type ProjectOverviewResponse struct {
	ID          uuid.UUID                        `json:"id"`
	Name        string                           `json:"name"`
	Description string                           `json:"description"`
	Slug        string                           `json:"slug"`
	RepoURL     string                           `json:"repo_url"`
	Status      models.ProjectStatus             `json:"status"`
	Customer    *ProjectOverviewCustomerResponse `json:"customer"`
	Services    []ProjectOverviewServiceResponse `json:"services"`
	CreatedAt   time.Time                        `json:"created_at"`
	UpdatedAt   time.Time                        `json:"updated_at"`
}

type ProjectOverviewCustomerResponse struct {
	ID                   uuid.UUID                   `json:"id"`
	Name                 string                      `json:"name"`
	Slug                 string                      `json:"slug"`
	Status               models.CustomerStatus       `json:"status"`
	ProjectValue         int                         `json:"project_value"`
	MonthlyValue         int                         `json:"monthly_value"`
	DueDay               int                         `json:"due_day"`
	ProjectPaymentStatus models.ProjectPaymentStatus `json:"project_payment_status"`
	LastPayment          *time.Time                  `json:"last_payment"`
}

type ProjectOverviewServiceResponse struct {
	ID             uuid.UUID            `json:"id"`
	Name           string               `json:"name"`
	Type           models.ProjectType   `json:"type"`
	URL            string               `json:"url"`
	RepoURL        string               `json:"repo_url"`
	Status         models.ProjectStatus `json:"status"`
	HealthCheckURL string               `json:"health_check_url"`
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

func ProjectOverviewResponseFromModel(project models.Project) ProjectOverviewResponse {
	services := make([]ProjectOverviewServiceResponse, 0, len(project.Services))
	for _, service := range project.Services {
		services = append(services, ProjectOverviewServiceResponse{
			ID:             service.ID,
			Name:           service.Name,
			Type:           service.Type,
			URL:            service.URL,
			RepoURL:        service.RepoURL,
			Status:         service.Status,
			HealthCheckURL: service.HealthCheckURL,
		})
	}

	return ProjectOverviewResponse{
		ID:          project.ID,
		Name:        project.Name,
		Description: project.Description,
		Slug:        project.Slug,
		RepoURL:     project.RepoURL,
		Status:      project.Status,
		Customer:    projectOverviewCustomerFromModel(project.CustomerProjects),
		Services:    services,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func projectOverviewCustomerFromModel(customerProjects []models.CustomerProject) *ProjectOverviewCustomerResponse {
	if len(customerProjects) == 0 {
		return nil
	}

	customerProject := customerProjects[0]
	return &ProjectOverviewCustomerResponse{
		ID:                   customerProject.Customer.ID,
		Name:                 customerProject.Customer.Name,
		Slug:                 customerProject.Customer.Slug,
		Status:               customerProject.Customer.Status,
		ProjectValue:         customerProject.ProjectValue,
		MonthlyValue:         customerProject.MonthlyValue,
		DueDay:               customerProject.DueDay,
		ProjectPaymentStatus: customerProject.ProjectPaymentStatus,
		LastPayment:          customerProject.LastPayment,
	}
}
