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
	Name        string `json:"name"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	RepoURL     string `json:"repo_url"`
}

type ProjectResponse struct {
	ID               uuid.UUID                        `json:"id"`
	Name             string                           `json:"name"`
	Description      string                           `json:"description"`
	Slug             string                           `json:"slug"`
	RepoURL          string                           `json:"repo_url"`
	Status           models.ProjectStatus             `json:"status"`
	CustomerProjects []ProjectCustomerProjectResponse `json:"customer_projects"`
	CreatedAt        time.Time                        `json:"created_at"`
	UpdatedAt        time.Time                        `json:"updated_at"`
}

type ProjectCustomerProjectResponse struct {
	ID            uuid.UUID                           `json:"id"`
	ProjectID     uuid.UUID                           `json:"project_id"`
	CustomerID    uuid.UUID                           `json:"customer_id"`
	Status        models.CustomerProjectStatus        `json:"status"`
	BillingStatus models.CustomerProjectBillingStatus `json:"billing_status"`
	ProjectValue  int                                 `json:"project_value"`
	MonthlyValue  int                                 `json:"monthly_value"`
	DueDay        int                                 `json:"due_day"`
	Customer      CustomerResponse                    `json:"customer"`
	StartedAt     time.Time                           `json:"started_at"`
	ClosedAt      *time.Time                          `json:"closed_at"`
	CreatedAt     time.Time                           `json:"created_at"`
	UpdatedAt     time.Time                           `json:"updated_at"`
}

type ProjectOverviewResponse struct {
	ID          uuid.UUID                         `json:"id"`
	Name        string                            `json:"name"`
	Description string                            `json:"description"`
	Slug        string                            `json:"slug"`
	RepoURL     string                            `json:"repo_url"`
	Status      models.ProjectStatus              `json:"status"`
	Customers   []ProjectOverviewCustomerResponse `json:"customers"`
	Services    []ProjectOverviewServiceResponse  `json:"services"`
	CreatedAt   time.Time                         `json:"created_at"`
	UpdatedAt   time.Time                         `json:"updated_at"`
}

type ProjectOverviewCustomerResponse struct {
	ID            uuid.UUID                           `json:"id"`
	Name          string                              `json:"name"`
	Slug          string                              `json:"slug"`
	Status        models.CustomerStatus               `json:"status"`
	ProjectValue  int                                 `json:"project_value"`
	MonthlyValue  int                                 `json:"monthly_value"`
	DueDay        int                                 `json:"due_day"`
	BillingStatus models.CustomerProjectBillingStatus `json:"billing_status"`
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

	if len(strings.TrimSpace(request.RepoURL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("PROJECT_REPO_URL_INVALID"))
	}

	return errs
}

func ProjectResponseFromModel(project models.Project) ProjectResponse {
	return ProjectResponse{
		ID:               project.ID,
		Name:             project.Name,
		Description:      project.Description,
		Slug:             project.Slug,
		RepoURL:          project.RepoURL,
		Status:           project.Status,
		CustomerProjects: projectCustomerProjectsFromModel(project.CustomerProjects),
		CreatedAt:        project.CreatedAt,
		UpdatedAt:        project.UpdatedAt,
	}
}

func projectCustomerProjectsFromModel(customerProjects []models.CustomerProject) []ProjectCustomerProjectResponse {
	items := make([]ProjectCustomerProjectResponse, 0, len(customerProjects))
	for _, customerProject := range customerProjects {
		projectValue, monthlyValue, dueDay := currentCommercialTerms(customerProject.Terms)
		items = append(items, ProjectCustomerProjectResponse{
			ID:            customerProject.ID,
			ProjectID:     customerProject.ProjectID,
			CustomerID:    customerProject.CustomerID,
			Status:        customerProject.Status,
			BillingStatus: CustomerProjectBillingStatusFromInvoices(customerProject.Status, customerProject.Invoices),
			ProjectValue:  projectValue,
			MonthlyValue:  monthlyValue,
			DueDay:        dueDay,
			Customer:      CustomerResponseFromModel(customerProject.Customer),
			StartedAt:     customerProject.StartedAt,
			ClosedAt:      customerProject.ClosedAt,
			CreatedAt:     customerProject.CreatedAt,
			UpdatedAt:     customerProject.UpdatedAt,
		})
	}

	return items
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
		Customers:   projectOverviewCustomersFromModel(project.CustomerProjects),
		Services:    services,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func projectOverviewCustomersFromModel(customerProjects []models.CustomerProject) []ProjectOverviewCustomerResponse {
	items := make([]ProjectOverviewCustomerResponse, 0, len(customerProjects))
	for _, customerProject := range customerProjects {
		projectValue, monthlyValue, dueDay := currentCommercialTerms(customerProject.Terms)
		items = append(items, ProjectOverviewCustomerResponse{
			ID:            customerProject.Customer.ID,
			Name:          customerProject.Customer.Name,
			Slug:          customerProject.Customer.Slug,
			Status:        customerProject.Customer.Status,
			ProjectValue:  projectValue,
			MonthlyValue:  monthlyValue,
			DueDay:        dueDay,
			BillingStatus: CustomerProjectBillingStatusFromInvoices(customerProject.Status, customerProject.Invoices),
		})
	}

	return items
}
