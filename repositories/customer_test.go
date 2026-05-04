package repositories

import (
	"context"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestNewCustomerRepository(t *testing.T) {
	repository := NewCustomerRepository(nil)
	if repository == nil {
		t.Fatal("Expected customer repository to be initialized")
	}
}

func TestCustomerRepositoryDryRunQueries(t *testing.T) {
	db := dryRunDB(t)
	repository := NewCustomerRepository(db)
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
		_, err := repository.Create(ctx, models.Customer{
			Name:         "TB Systems",
			Slug:         "tb-systems",
			Document:     "12345678000199",
			DocumentType: models.CustomerDocumentTypeCNPJ,
			Email:        "contact@tbsystems.com.br",
			Status:       models.CustomerStatusActive,
		})
		if err != nil {
			t.Fatalf("Expected dry run create to succeed, got %v", err)
		}
	})

	t.Run("update builds query", func(t *testing.T) {
		_, err := repository.Update(ctx, models.Customer{
			ID:           id,
			Name:         "TB Systems",
			Slug:         "tb-systems",
			Document:     "12345678000199",
			DocumentType: models.CustomerDocumentTypeCNPJ,
			Email:        "contact@tbsystems.com.br",
			Status:       models.CustomerStatusOnboarding,
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

	t.Run("unique lookups build queries", func(t *testing.T) {
		checks := []struct {
			name string
			run  func() (bool, error)
		}{
			{name: "slug", run: func() (bool, error) { return repository.SlugExists(ctx, "tb-systems", &id) }},
			{name: "document", run: func() (bool, error) { return repository.DocumentExists(ctx, "12345678000199", &id) }},
			{name: "email", run: func() (bool, error) { return repository.EmailExists(ctx, "contact@tbsystems.com.br", &id) }},
		}

		for _, check := range checks {
			exists, err := check.run()
			if err != nil {
				t.Fatalf("Expected dry run %s lookup to succeed, got %v", check.name, err)
			}
			if exists {
				t.Fatalf("Expected dry run %s lookup to return false", check.name)
			}
		}
	})
}
