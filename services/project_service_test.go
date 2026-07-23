package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
)

func TestProjectServiceListFiltersByProjectID(t *testing.T) {
	projectID := uuid.New()
	repository := &fakeProjectServiceRepository{
		projectServices: []models.ProjectService{
			{Name: "API", ProjectID: projectID},
		},
		total: 1,
	}
	service := NewProjectServiceService(repository, &fakeProjectServiceStatusSync{})

	response, apiErr := service.List(
		context.Background(),
		commonsmodels.PaginatedParams{Limit: 10, Page: 1},
		dto.ProjectServiceListFilter{ProjectID: projectID.String()},
	)

	if apiErr != nil {
		t.Fatalf("Expected project services list, got status %d", apiErr.GetStatus())
	}
	if repository.receivedFilter.ProjectID == nil || *repository.receivedFilter.ProjectID != projectID {
		t.Fatalf("Expected project id filter %s, got %#v", projectID, repository.receivedFilter.ProjectID)
	}
	if len(response.Items) != 1 {
		t.Fatalf("Expected 1 project service, got %d", len(response.Items))
	}
}

func TestProjectServiceListRejectsInvalidProjectIDFilter(t *testing.T) {
	service := NewProjectServiceService(&fakeProjectServiceRepository{}, &fakeProjectServiceStatusSync{})

	_, apiErr := service.List(
		context.Background(),
		commonsmodels.PaginatedParams{Limit: 10, Page: 1},
		dto.ProjectServiceListFilter{ProjectID: "invalid"},
	)

	if apiErr == nil {
		t.Fatalf("Expected invalid project id error")
	}
	if apiErr.GetStatus() != http.StatusBadRequest {
		t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, apiErr.GetStatus())
	}
}

func TestProjectServiceCreateSyncsProjectStatus(t *testing.T) {
	projectID := uuid.New()
	statusSync := &fakeProjectServiceStatusSync{}
	service := NewProjectServiceService(&fakeProjectServiceRepository{}, statusSync)

	_, apiErr := service.Create(context.Background(), dto.ProjectServiceCreateRequest{
		ProjectID: projectID,
		Name:      "API",
		Type:      models.ProjectTypeBackend,
	})

	if apiErr != nil {
		t.Fatalf("Expected project service creation, got status %d", apiErr.GetStatus())
	}
	if len(statusSync.syncedProjectIDs) != 1 || statusSync.syncedProjectIDs[0] != projectID {
		t.Fatalf("Expected project %s to be synced, got %#v", projectID, statusSync.syncedProjectIDs)
	}
}

func TestProjectServiceCreateStartsAsBacklog(t *testing.T) {
	repository := &fakeProjectServiceRepository{}
	service := NewProjectServiceService(repository, &fakeProjectServiceStatusSync{})

	response, apiErr := service.Create(context.Background(), dto.ProjectServiceCreateRequest{
		ProjectID: uuid.New(),
		Name:      "API",
		Type:      models.ProjectTypeBackend,
	})

	if apiErr != nil {
		t.Fatalf("Expected project service creation, got status %d", apiErr.GetStatus())
	}
	if response.Status != models.ProjectStatusBacklog {
		t.Fatalf("Expected created project service status Backlog, got %d", response.Status)
	}
	if repository.createdProjectService.Status != models.ProjectStatusBacklog {
		t.Fatalf("Expected repository project service status Backlog, got %d", repository.createdProjectService.Status)
	}
}

func TestProjectServiceUpdateSyncsOldAndNewProjectStatus(t *testing.T) {
	oldProjectID := uuid.New()
	newProjectID := uuid.New()
	serviceID := uuid.New()
	statusSync := &fakeProjectServiceStatusSync{}
	service := NewProjectServiceService(&fakeProjectServiceRepository{
		projectService: models.ProjectService{ID: serviceID, ProjectID: oldProjectID},
	}, statusSync)

	_, apiErr := service.Update(context.Background(), serviceID.String(), dto.ProjectServiceRequest{
		ProjectID: newProjectID,
		Name:      "API",
		Type:      models.ProjectTypeBackend,
		Status:    models.ProjectStatusLive,
	})

	if apiErr != nil {
		t.Fatalf("Expected project service update, got status %d", apiErr.GetStatus())
	}
	if len(statusSync.syncedProjectIDs) != 2 {
		t.Fatalf("Expected old and new project sync, got %#v", statusSync.syncedProjectIDs)
	}
	if statusSync.syncedProjectIDs[0] != oldProjectID || statusSync.syncedProjectIDs[1] != newProjectID {
		t.Fatalf("Expected sync order [%s %s], got %#v", oldProjectID, newProjectID, statusSync.syncedProjectIDs)
	}
}

func TestProjectServiceDeleteSyncsProjectStatus(t *testing.T) {
	projectID := uuid.New()
	serviceID := uuid.New()
	statusSync := &fakeProjectServiceStatusSync{}
	service := NewProjectServiceService(&fakeProjectServiceRepository{
		projectService: models.ProjectService{ID: serviceID, ProjectID: projectID},
	}, statusSync)

	apiErr := service.Delete(context.Background(), serviceID.String())

	if apiErr != nil {
		t.Fatalf("Expected project service deletion, got status %d", apiErr.GetStatus())
	}
	if len(statusSync.syncedProjectIDs) != 1 || statusSync.syncedProjectIDs[0] != projectID {
		t.Fatalf("Expected project %s to be synced, got %#v", projectID, statusSync.syncedProjectIDs)
	}
}

type fakeProjectServiceRepository struct {
	projectServices       []models.ProjectService
	projectService        models.ProjectService
	createdProjectService models.ProjectService
	receivedFilter        repositories.ProjectServiceFilter
	total                 int64
}

func (f *fakeProjectServiceRepository) List(_ context.Context, _ commonsmodels.PaginatedParams, filter repositories.ProjectServiceFilter) ([]models.ProjectService, int64, error) {
	f.receivedFilter = filter
	return f.projectServices, f.total, nil
}

func (f *fakeProjectServiceRepository) FindByID(_ context.Context, id uuid.UUID) (models.ProjectService, error) {
	if f.projectService.ID == uuid.Nil {
		f.projectService.ID = id
	}
	return f.projectService, nil
}

func (f *fakeProjectServiceRepository) Create(_ context.Context, projectService models.ProjectService) (models.ProjectService, error) {
	f.createdProjectService = projectService
	return projectService, nil
}

func (f *fakeProjectServiceRepository) Update(_ context.Context, projectService models.ProjectService) (models.ProjectService, error) {
	return projectService, nil
}

func (f *fakeProjectServiceRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

type fakeProjectServiceStatusSync struct {
	syncedProjectIDs []uuid.UUID
}

func (f *fakeProjectServiceStatusSync) Sync(_ context.Context, projectID uuid.UUID) errors.ApiError {
	f.syncedProjectIDs = append(f.syncedProjectIDs, projectID)
	return nil
}
