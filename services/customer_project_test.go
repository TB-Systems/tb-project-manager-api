package services

import (
	"context"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestCustomerProjectServiceCreate(t *testing.T) {
	t.Run("creates link when project can have more than one customer", func(t *testing.T) {
		repository := &fakeCustomerProjectRepository{}
		statusSync := &fakeCustomerProjectStatusSync{}
		service := NewCustomerProjectService(repository, statusSync)

		response, apiErr := service.Create(context.Background(), validCustomerProjectRequest())

		if apiErr != nil {
			t.Fatalf("Expected customer project to be created, got status %d", apiErr.GetStatus())
		}
		if response.ProjectID == uuid.Nil {
			t.Fatal("Expected project ID in response")
		}
		if len(statusSync.syncedProjectIDs) != 1 || statusSync.syncedProjectIDs[0] != response.ProjectID {
			t.Fatalf("Expected linked project %s status to be synced, got %#v", response.ProjectID, statusSync.syncedProjectIDs)
		}
		if response.Status != models.CustomerProjectStatusActive {
			t.Fatalf("Expected active customer project status, got %d", response.Status)
		}
		if len(response.Terms) != 1 {
			t.Fatalf("Expected active commercial term to be created, got %d", len(response.Terms))
		}
		if len(response.Invoices) != 2 {
			t.Fatalf("Expected setup invoices to be created, got %d", len(response.Invoices))
		}
		if response.BillingStatus != models.CustomerProjectBillingStatusSetupPending {
			t.Fatalf("Expected setup pending billing status, got %d", response.BillingStatus)
		}
	})
}

func TestCustomerProjectServiceUpdate(t *testing.T) {
	t.Run("syncs old and new project when moving customer link", func(t *testing.T) {
		id := uuid.New()
		oldProjectID := uuid.New()
		newProjectID := uuid.New()
		statusSync := &fakeCustomerProjectStatusSync{}
		repository := &fakeCustomerProjectRepository{
			customerProject: models.CustomerProject{ID: id, ProjectID: oldProjectID},
		}
		service := NewCustomerProjectService(repository, statusSync)
		request := validCustomerProjectRequest()
		request.ProjectID = newProjectID

		_, apiErr := service.Update(context.Background(), id.String(), request)

		if apiErr != nil {
			t.Fatalf("Expected customer project update, got status %d", apiErr.GetStatus())
		}
		if len(statusSync.syncedProjectIDs) != 2 {
			t.Fatalf("Expected old and new project sync, got %#v", statusSync.syncedProjectIDs)
		}
		if statusSync.syncedProjectIDs[0] != oldProjectID || statusSync.syncedProjectIDs[1] != newProjectID {
			t.Fatalf("Expected sync order [%s %s], got %#v", oldProjectID, newProjectID, statusSync.syncedProjectIDs)
		}
	})
}

func TestCustomerProjectServiceDelete(t *testing.T) {
	id := uuid.New()
	projectID := uuid.New()
	statusSync := &fakeCustomerProjectStatusSync{}
	repository := &fakeCustomerProjectRepository{
		customerProject: models.CustomerProject{ID: id, ProjectID: projectID},
	}
	service := NewCustomerProjectService(repository, statusSync)

	apiErr := service.Delete(context.Background(), id.String())

	if apiErr != nil {
		t.Fatalf("Expected customer project deletion, got status %d", apiErr.GetStatus())
	}
	if len(statusSync.syncedProjectIDs) != 1 || statusSync.syncedProjectIDs[0] != projectID {
		t.Fatalf("Expected project %s to be synced, got %#v", projectID, statusSync.syncedProjectIDs)
	}
}

func validCustomerProjectRequest() dto.CustomerProjectRequest {
	return dto.CustomerProjectRequest{
		ProjectID:    uuid.New(),
		CustomerID:   uuid.New(),
		ProjectValue: 10000,
		MonthlyValue: 1500,
		DueDay:       10,
	}
}

type fakeCustomerProjectRepository struct {
	customerProject models.CustomerProject
	findErr         error
	linkExists      bool
	linkErr         error
}

func (f *fakeCustomerProjectRepository) List(context.Context, commonsmodels.PaginatedParams) ([]models.CustomerProject, int64, error) {
	return nil, 0, nil
}

func (f *fakeCustomerProjectRepository) FindByID(_ context.Context, id uuid.UUID) (models.CustomerProject, error) {
	if f.findErr != nil {
		return models.CustomerProject{}, f.findErr
	}
	if f.customerProject.ID == uuid.Nil {
		f.customerProject.ID = id
	}
	return f.customerProject, nil
}

func (f *fakeCustomerProjectRepository) Create(_ context.Context, customerProject models.CustomerProject) (models.CustomerProject, error) {
	return customerProject, nil
}

func (f *fakeCustomerProjectRepository) Update(_ context.Context, customerProject models.CustomerProject) (models.CustomerProject, error) {
	return customerProject, nil
}

func (f *fakeCustomerProjectRepository) Delete(context.Context, uuid.UUID) error {
	return nil
}

func (f *fakeCustomerProjectRepository) LinkExists(context.Context, uuid.UUID, uuid.UUID, *uuid.UUID) (bool, error) {
	if f.linkErr != nil {
		return false, f.linkErr
	}
	return f.linkExists, nil
}

type fakeCustomerProjectStatusSync struct {
	syncedProjectIDs []uuid.UUID
}

func (f *fakeCustomerProjectStatusSync) Sync(_ context.Context, projectID uuid.UUID) errors.ApiError {
	f.syncedProjectIDs = append(f.syncedProjectIDs, projectID)
	return nil
}
