package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestProjectServiceCreate(t *testing.T) {
	t.Run("creates project with trimmed fields", func(t *testing.T) {
		repository := &fakeProjectRepository{}
		service := NewProjectService(repository)

		response, apiErr := service.Create(context.Background(), dto.ProjectRequest{
			Name:        " TB Manager ",
			Description: " Internal manager ",
			Slug:        " tb-manager ",
			RepoURL:     " https://github.com/TB-Systems/tb-manager ",
		})

		if apiErr != nil {
			t.Fatalf("Expected project to be created, got status %d", apiErr.GetStatus())
		}
		if repository.createdProject.Name != "TB Manager" {
			t.Fatalf("Expected trimmed project name, got %q", repository.createdProject.Name)
		}
		if response.Slug != "tb-manager" {
			t.Fatalf("Expected created response slug, got %q", response.Slug)
		}
		if repository.createdProject.Description != "Internal manager" {
			t.Fatalf("Expected trimmed project description, got %q", repository.createdProject.Description)
		}
		if repository.createdProject.Status != models.ProjectStatusBacklog {
			t.Fatalf("Expected created project status Backlog, got %d", repository.createdProject.Status)
		}
	})

	t.Run("rejects duplicated slug", func(t *testing.T) {
		repository := &fakeProjectRepository{slugExists: true}
		service := NewProjectService(repository)

		_, apiErr := service.Create(context.Background(), dto.ProjectRequest{
			Name: "TB Manager",
			Slug: "tb-manager",
		})

		if apiErr == nil {
			t.Fatal("Expected duplicated slug error")
		}
		if apiErr.GetStatus() != http.StatusConflict {
			t.Fatalf("Expected status %d, got %d", http.StatusConflict, apiErr.GetStatus())
		}
	})
}

func TestProjectServiceFindByID(t *testing.T) {
	t.Run("rejects invalid id", func(t *testing.T) {
		service := NewProjectService(&fakeProjectRepository{})

		_, apiErr := service.FindByID(context.Background(), "invalid")

		if apiErr == nil {
			t.Fatal("Expected invalid ID error")
		}
		if apiErr.GetStatus() != http.StatusBadRequest {
			t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, apiErr.GetStatus())
		}
	})

	t.Run("returns not found", func(t *testing.T) {
		service := NewProjectService(&fakeProjectRepository{findErr: gorm.ErrRecordNotFound})

		_, apiErr := service.FindByID(context.Background(), uuid.NewString())

		if apiErr == nil {
			t.Fatal("Expected not found error")
		}
		if apiErr.GetStatus() != http.StatusNotFound {
			t.Fatalf("Expected status %d, got %d", http.StatusNotFound, apiErr.GetStatus())
		}
	})

	t.Run("returns customer project relation when project has customer", func(t *testing.T) {
		customerProjectID := uuid.New()
		customerID := uuid.New()
		projectID := uuid.New()
		service := NewProjectService(&fakeProjectRepository{
			project: models.Project{
				ID: projectID,
				CustomerProjects: []models.CustomerProject{
					{
						ID:                   customerProjectID,
						ProjectID:            projectID,
						CustomerID:           customerID,
						ProjectValue:         250000,
						MonthlyValue:         15000,
						DueDay:               10,
						ProjectPaymentStatus: models.ProjectPaymentStatusFirstHalfPending,
						Customer: models.Customer{
							ID:     customerID,
							Name:   "TiB",
							Slug:   "tib",
							Status: models.CustomerStatusActive,
						},
					},
				},
			},
		})

		response, apiErr := service.FindByID(context.Background(), projectID.String())

		if apiErr != nil {
			t.Fatalf("Expected project detail, got status %d", apiErr.GetStatus())
		}
		if response.CustomerProject == nil {
			t.Fatal("Expected customer project relation")
		}
		if response.CustomerProject.ID != customerProjectID {
			t.Fatalf("Expected customer project ID %s, got %s", customerProjectID, response.CustomerProject.ID)
		}
		if response.CustomerProject.Customer.ID != customerID {
			t.Fatalf("Expected nested customer ID %s, got %s", customerID, response.CustomerProject.Customer.ID)
		}
	})
}

func TestProjectServiceList(t *testing.T) {
	repository := &fakeProjectRepository{
		projects: []models.Project{
			{Name: "Project 1", Slug: "project-1"},
			{Name: "Project 2", Slug: "project-2"},
		},
		total: 12,
	}
	service := NewProjectService(repository)

	response, apiErr := service.List(context.Background(), commonsmodels.PaginatedParams{Limit: 5, Page: 2})

	if apiErr != nil {
		t.Fatalf("Expected projects list, got status %d", apiErr.GetStatus())
	}
	if len(response.Items) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(response.Items))
	}
	if response.PageCount != 3 {
		t.Fatalf("Expected page count 3, got %d", response.PageCount)
	}
}

func TestProjectServiceOverview(t *testing.T) {
	customerID := uuid.New()
	repository := &fakeProjectRepository{
		projects: []models.Project{
			{
				Name:   "Project with customer",
				Slug:   "project-with-customer",
				Status: models.ProjectStatusDeveloping,
				CustomerProjects: []models.CustomerProject{
					{
						ProjectValue:         10000,
						MonthlyValue:         1500,
						DueDay:               10,
						ProjectPaymentStatus: models.ProjectPaymentStatusFirstHalfPaid,
						Customer: models.Customer{
							ID:     customerID,
							Name:   "ACME",
							Slug:   "acme",
							Status: models.CustomerStatusActive,
						},
					},
				},
				Services: []models.ProjectService{
					{Name: "API", Type: models.ProjectTypeBackend, Status: models.ProjectStatusDeveloping},
				},
			},
			{
				Name:   "Project without customer",
				Slug:   "project-without-customer",
				Status: models.ProjectStatusBacklog,
			},
		},
		total: 2,
	}
	service := NewProjectService(repository)

	response, apiErr := service.Overview(context.Background())

	if apiErr != nil {
		t.Fatalf("Expected project overview, got status %d", apiErr.GetStatus())
	}
	if len(response.Items) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(response.Items))
	}
	if response.Items[0].Customer == nil {
		t.Fatal("Expected first project customer")
	}
	if response.Items[0].Customer.ID != customerID {
		t.Fatalf("Expected customer ID %s, got %s", customerID, response.Items[0].Customer.ID)
	}
	if len(response.Items[0].Services) != 1 {
		t.Fatalf("Expected one service, got %d", len(response.Items[0].Services))
	}
	if response.Items[1].Customer != nil {
		t.Fatal("Expected second project without customer")
	}
	if response.Total != 2 {
		t.Fatalf("Expected total 2, got %d", response.Total)
	}
}

func TestProjectServiceUpdateAndDelete(t *testing.T) {
	t.Run("updates existing project", func(t *testing.T) {
		id := uuid.New()
		repository := &fakeProjectRepository{project: models.Project{ID: id, Status: models.ProjectStatusStaging}}
		service := NewProjectService(repository)

		response, apiErr := service.Update(context.Background(), id.String(), dto.ProjectRequest{
			Name: "TB Manager",
			Slug: "tb-manager",
		})

		if apiErr != nil {
			t.Fatalf("Expected project update, got status %d", apiErr.GetStatus())
		}
		if response.Status != models.ProjectStatusStaging {
			t.Fatalf("Expected existing project status to be preserved, got %d", response.Status)
		}
	})

	t.Run("deletes existing project", func(t *testing.T) {
		id := uuid.New()
		repository := &fakeProjectRepository{project: models.Project{ID: id}}
		service := NewProjectService(repository)

		apiErr := service.Delete(context.Background(), id.String())

		if apiErr != nil {
			t.Fatalf("Expected project deletion, got status %d", apiErr.GetStatus())
		}
		if repository.deletedID != id {
			t.Fatalf("Expected deleted ID %s, got %s", id, repository.deletedID)
		}
	})
}

type fakeProjectRepository struct {
	createdProject models.Project
	deletedID      uuid.UUID
	findErr        error
	project        models.Project
	projects       []models.Project
	slugErr        error
	slugExists     bool
	total          int64
	updateErr      error
}

func (f *fakeProjectRepository) List(context.Context, commonsmodels.PaginatedParams) ([]models.Project, int64, error) {
	return f.projects, f.total, nil
}

func (f *fakeProjectRepository) Overview(context.Context) ([]models.Project, error) {
	return f.projects, nil
}

func (f *fakeProjectRepository) FindByID(_ context.Context, id uuid.UUID) (models.Project, error) {
	if f.findErr != nil {
		return models.Project{}, f.findErr
	}
	if f.project.ID == uuid.Nil {
		f.project.ID = id
	}
	return f.project, nil
}

func (f *fakeProjectRepository) Create(_ context.Context, project models.Project) (models.Project, error) {
	f.createdProject = project
	return project, nil
}

func (f *fakeProjectRepository) Update(_ context.Context, project models.Project) (models.Project, error) {
	if f.updateErr != nil {
		return models.Project{}, f.updateErr
	}
	return project, nil
}

func (f *fakeProjectRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deletedID = id
	return nil
}

func (f *fakeProjectRepository) SlugExists(context.Context, string, *uuid.UUID) (bool, error) {
	if f.slugErr != nil {
		return false, f.slugErr
	}
	return f.slugExists, nil
}

var errProjectRepository = stderrors.New("project repository error")
