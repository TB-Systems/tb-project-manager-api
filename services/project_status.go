package services

import (
	"context"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
)

type ProjectStatusSync interface {
	Sync(ctx context.Context, projectID uuid.UUID) errors.ApiError
}

type projectStatusSync struct {
	repository repositories.ProjectStatus
}

func NewProjectStatusSyncService(repository repositories.ProjectStatus) ProjectStatusSync {
	return projectStatusSync{repository: repository}
}

func (p projectStatusSync) Sync(ctx context.Context, projectID uuid.UUID) errors.ApiError {
	hasCustomer, err := p.repository.HasCustomerProject(ctx, projectID)
	if err != nil {
		return internalProjectError("CHECK_PROJECT_CUSTOMER_LINK_FAILED")
	}

	services, err := p.repository.ListProjectServices(ctx, projectID)
	if err != nil {
		return internalProjectError("LIST_PROJECT_SERVICES_FOR_STATUS_FAILED")
	}

	status := ResolveProjectStatus(hasCustomer, services)
	if err := p.repository.UpdateProjectStatus(ctx, projectID, status); err != nil {
		return internalProjectError("UPDATE_PROJECT_STATUS_FAILED")
	}

	return nil
}

func ResolveProjectStatus(hasCustomer bool, services []models.ProjectService) models.ProjectStatus {
	if len(services) == 0 {
		if hasCustomer {
			return models.ProjectStatusDiscovery
		}

		return models.ProjectStatusBacklog
	}

	if allServicesHaveStatus(services, models.ProjectStatusPaused) {
		return models.ProjectStatusPaused
	}

	if allServicesHaveStatus(services, models.ProjectStatusArchived) {
		return models.ProjectStatusArchived
	}

	for _, status := range []models.ProjectStatus{
		models.ProjectStatusDown,
		models.ProjectStatusLive,
		models.ProjectStatusStaging,
	} {
		if anyServiceHasStatus(services, status) {
			return status
		}
	}

	for _, status := range []models.ProjectStatus{
		models.ProjectStatusDeveloping,
		models.ProjectStatusDiscovery,
		models.ProjectStatusBacklog,
	} {
		if anyServiceHasStatus(services, status) {
			return models.ProjectStatusDeveloping
		}
	}

	return models.ProjectStatusDeveloping
}

func allServicesHaveStatus(services []models.ProjectService, status models.ProjectStatus) bool {
	for _, service := range services {
		if service.Status != status {
			return false
		}
	}

	return true
}

func anyServiceHasStatus(services []models.ProjectService, status models.ProjectStatus) bool {
	for _, service := range services {
		if service.Status == status {
			return true
		}
	}

	return false
}
