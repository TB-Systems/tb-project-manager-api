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
			Name:   " TB Manager ",
			Slug:   " tb-manager ",
			Type:   models.ProjectTypeBackend,
			Status: models.ProjectStatusBacklog,
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
	})

	t.Run("rejects duplicated slug", func(t *testing.T) {
		repository := &fakeProjectRepository{slugExists: true}
		service := NewProjectService(repository)

		_, apiErr := service.Create(context.Background(), dto.ProjectRequest{
			Name:   "TB Manager",
			Slug:   "tb-manager",
			Type:   models.ProjectTypeBackend,
			Status: models.ProjectStatusBacklog,
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

func TestProjectServiceUpdateAndDelete(t *testing.T) {
	t.Run("updates existing project", func(t *testing.T) {
		id := uuid.New()
		repository := &fakeProjectRepository{project: models.Project{ID: id}}
		service := NewProjectService(repository)

		response, apiErr := service.Update(context.Background(), id.String(), dto.ProjectRequest{
			Name:   "TB Manager",
			Slug:   "tb-manager",
			Type:   models.ProjectTypeBackend,
			Status: models.ProjectStatusDeveloping,
		})

		if apiErr != nil {
			t.Fatalf("Expected project update, got status %d", apiErr.GetStatus())
		}
		if response.Status != models.ProjectStatusDeveloping {
			t.Fatalf("Expected updated project status, got %d", response.Status)
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
