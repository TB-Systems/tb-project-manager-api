package repositories

import (
	"context"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestNewProjectRepository(t *testing.T) {
	repository := NewProjectRepository(nil)
	if repository == nil {
		t.Fatal("Expected project repository to be initialized")
	}
}

func TestProjectRepositoryDryRunQueries(t *testing.T) {
	db := dryRunDB(t)
	repository := NewProjectRepository(db)
	ctx := context.Background()
	id := uuid.New()

	t.Run("list builds query", func(t *testing.T) {
		_, _, err := repository.List(ctx, commonsmodels.PaginatedParams{Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("Expected dry run list to succeed, got %v", err)
		}
	})

	t.Run("find by id builds query", func(t *testing.T) {
		_, err := repository.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("Expected dry run find to succeed, got %v", err)
		}
	})

	t.Run("create builds query", func(t *testing.T) {
		_, err := repository.Create(ctx, models.Project{
			Name:   "TB Manager",
			Slug:   "tb-manager",
			Status: models.ProjectStatusBacklog,
		})
		if err != nil {
			t.Fatalf("Expected dry run create to succeed, got %v", err)
		}
	})

	t.Run("update builds query", func(t *testing.T) {
		_, err := repository.Update(ctx, models.Project{
			ID:     id,
			Name:   "TB Manager",
			Slug:   "tb-manager",
			Status: models.ProjectStatusDeveloping,
		})
		if err != nil {
			t.Fatalf("Expected dry run update to succeed, got %v", err)
		}
	})

	t.Run("delete builds query", func(t *testing.T) {
		err := repository.Delete(ctx, id)
		if err != nil {
			t.Fatalf("Expected dry run delete to succeed, got %v", err)
		}
	})

	t.Run("slug exists builds query", func(t *testing.T) {
		exists, err := repository.SlugExists(ctx, "tb-manager", &id)
		if err != nil {
			t.Fatalf("Expected dry run slug lookup to succeed, got %v", err)
		}
		if exists {
			t.Fatal("Expected dry run slug lookup to return false")
		}
	})
}
