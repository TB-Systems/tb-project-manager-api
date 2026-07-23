package services

import (
	"context"
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestResolveProjectStatus(t *testing.T) {
	tests := []struct {
		name        string
		hasCustomer bool
		services    []models.ProjectService
		expected    models.ProjectStatus
	}{
		{
			name:     "sem cliente e sem servicos fica em backlog",
			expected: models.ProjectStatusBacklog,
		},
		{
			name:        "com cliente e sem servicos fica em discovery",
			hasCustomer: true,
			expected:    models.ProjectStatusDiscovery,
		},
		{
			name:        "developing sobe projeto se nao houver status maior",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusDeveloping, models.ProjectStatusBacklog),
			expected:    models.ProjectStatusDeveloping,
		},
		{
			name:        "servico em backlog deixa projeto em developing",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusBacklog),
			expected:    models.ProjectStatusDeveloping,
		},
		{
			name:        "servico em discovery deixa projeto em developing",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusDiscovery),
			expected:    models.ProjectStatusDeveloping,
		},
		{
			name:        "staging tem prioridade sobre developing",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusDeveloping, models.ProjectStatusStaging),
			expected:    models.ProjectStatusStaging,
		},
		{
			name:        "live tem prioridade sobre staging",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusLive, models.ProjectStatusStaging),
			expected:    models.ProjectStatusLive,
		},
		{
			name:        "down tem prioridade sobre qualquer status",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusDown, models.ProjectStatusLive),
			expected:    models.ProjectStatusDown,
		},
		{
			name:        "paused so vale quando todos os servicos estao paused",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusPaused, models.ProjectStatusPaused),
			expected:    models.ProjectStatusPaused,
		},
		{
			name:        "paused misturado nao vence live",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusPaused, models.ProjectStatusLive),
			expected:    models.ProjectStatusLive,
		},
		{
			name:        "archived so vale quando todos os servicos estao archived",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusArchived, models.ProjectStatusArchived),
			expected:    models.ProjectStatusArchived,
		},
		{
			name:        "archived misturado nao vence developing",
			hasCustomer: true,
			services:    projectServicesWithStatuses(models.ProjectStatusArchived, models.ProjectStatusDeveloping),
			expected:    models.ProjectStatusDeveloping,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ResolveProjectStatus(tt.hasCustomer, tt.services)
			if got != tt.expected {
				t.Fatalf("Expected status %d, got %d", tt.expected, got)
			}
		})
	}
}

func TestProjectStatusSync(t *testing.T) {
	projectID := uuid.New()
	repository := &fakeProjectStatusRepository{
		hasCustomer: true,
		services:    projectServicesWithStatuses(models.ProjectStatusLive),
	}
	service := NewProjectStatusSyncService(repository)

	apiErr := service.Sync(context.Background(), projectID)

	if apiErr != nil {
		t.Fatalf("Expected project status sync, got status %d", apiErr.GetStatus())
	}
	if repository.updatedProjectID != projectID {
		t.Fatalf("Expected project %s to be updated, got %s", projectID, repository.updatedProjectID)
	}
	if repository.updatedStatus != models.ProjectStatusLive {
		t.Fatalf("Expected project status Live, got %d", repository.updatedStatus)
	}
}

func projectServicesWithStatuses(statuses ...models.ProjectStatus) []models.ProjectService {
	services := make([]models.ProjectService, 0, len(statuses))
	for _, status := range statuses {
		services = append(services, models.ProjectService{Status: status})
	}

	return services
}

type fakeProjectStatusRepository struct {
	hasCustomer      bool
	services         []models.ProjectService
	updatedProjectID uuid.UUID
	updatedStatus    models.ProjectStatus
}

func (f *fakeProjectStatusRepository) HasCustomerProject(context.Context, uuid.UUID) (bool, error) {
	return f.hasCustomer, nil
}

func (f *fakeProjectStatusRepository) ListProjectServices(context.Context, uuid.UUID) ([]models.ProjectService, error) {
	return f.services, nil
}

func (f *fakeProjectStatusRepository) UpdateProjectStatus(_ context.Context, projectID uuid.UUID, status models.ProjectStatus) error {
	f.updatedProjectID = projectID
	f.updatedStatus = status
	return nil
}
