package services

import (
	"context"
	"net/http"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCustomerServiceCreate(t *testing.T) {
	t.Run("creates customer with trimmed fields", func(t *testing.T) {
		repository := &fakeCustomerRepository{}
		service := NewCustomerService(repository)

		response, apiErr := service.Create(context.Background(), dto.CustomerRequest{
			Name:         " TB Systems ",
			Slug:         " tb-systems ",
			Document:     " 04.252.011/0001-10 ",
			DocumentType: models.CustomerDocumentTypeCNPJ,
			Email:        " contact@tbsystems.com.br ",
			Status:       models.CustomerStatusActive,
		})

		if apiErr != nil {
			t.Fatalf("Expected customer to be created, got status %d", apiErr.GetStatus())
		}
		if repository.createdCustomer.Name != "TB Systems" {
			t.Fatalf("Expected trimmed customer name, got %q", repository.createdCustomer.Name)
		}
		if repository.createdCustomer.Document != "04252011000110" {
			t.Fatalf("Expected normalized customer document, got %q", repository.createdCustomer.Document)
		}
		if response.Email != "contact@tbsystems.com.br" {
			t.Fatalf("Expected created response email, got %q", response.Email)
		}
	})

	t.Run("rejects duplicated slug", func(t *testing.T) {
		service := NewCustomerService(&fakeCustomerRepository{slugExists: true})

		_, apiErr := service.Create(context.Background(), validCustomerRequest())

		assertStatus(t, apiErr, http.StatusConflict)
	})

	t.Run("rejects duplicated document", func(t *testing.T) {
		service := NewCustomerService(&fakeCustomerRepository{documentExists: true})

		_, apiErr := service.Create(context.Background(), validCustomerRequest())

		assertStatus(t, apiErr, http.StatusConflict)
	})

	t.Run("rejects duplicated email", func(t *testing.T) {
		service := NewCustomerService(&fakeCustomerRepository{emailExists: true})

		_, apiErr := service.Create(context.Background(), validCustomerRequest())

		assertStatus(t, apiErr, http.StatusConflict)
	})
}

func TestCustomerServiceFindByID(t *testing.T) {
	t.Run("rejects invalid id", func(t *testing.T) {
		service := NewCustomerService(&fakeCustomerRepository{})

		_, apiErr := service.FindByID(context.Background(), "invalid")

		assertStatus(t, apiErr, http.StatusBadRequest)
	})

	t.Run("returns not found", func(t *testing.T) {
		service := NewCustomerService(&fakeCustomerRepository{findErr: gorm.ErrRecordNotFound})

		_, apiErr := service.FindByID(context.Background(), uuid.NewString())

		assertStatus(t, apiErr, http.StatusNotFound)
	})
}

func TestCustomerServiceList(t *testing.T) {
	repository := &fakeCustomerRepository{
		customers: []models.Customer{
			{Name: "Customer 1", Slug: "customer-1"},
			{Name: "Customer 2", Slug: "customer-2"},
		},
		total: 12,
	}
	service := NewCustomerService(repository)

	response, apiErr := service.List(context.Background(), commonsmodels.PaginatedParams{Limit: 5, Page: 2})

	if apiErr != nil {
		t.Fatalf("Expected customers list, got status %d", apiErr.GetStatus())
	}
	if len(response.Items) != 2 {
		t.Fatalf("Expected 2 customers, got %d", len(response.Items))
	}
	if response.PageCount != 3 {
		t.Fatalf("Expected page count 3, got %d", response.PageCount)
	}
}

func TestCustomerServiceUpdateAndDelete(t *testing.T) {
	t.Run("updates existing customer", func(t *testing.T) {
		id := uuid.New()
		repository := &fakeCustomerRepository{customer: models.Customer{ID: id}}
		service := NewCustomerService(repository)

		response, apiErr := service.Update(context.Background(), id.String(), validCustomerRequest())

		if apiErr != nil {
			t.Fatalf("Expected customer update, got status %d", apiErr.GetStatus())
		}
		if response.Status != models.CustomerStatusActive {
			t.Fatalf("Expected updated customer status, got %d", response.Status)
		}
	})

	t.Run("deletes existing customer", func(t *testing.T) {
		id := uuid.New()
		repository := &fakeCustomerRepository{customer: models.Customer{ID: id}}
		service := NewCustomerService(repository)

		apiErr := service.Delete(context.Background(), id.String())

		if apiErr != nil {
			t.Fatalf("Expected customer deletion, got status %d", apiErr.GetStatus())
		}
		if repository.deletedID != id {
			t.Fatalf("Expected deleted ID %s, got %s", id, repository.deletedID)
		}
	})
}

func validCustomerRequest() dto.CustomerRequest {
	return dto.CustomerRequest{
		Name:         "TB Systems",
		Slug:         "tb-systems",
		Document:     "04.252.011/0001-10",
		DocumentType: models.CustomerDocumentTypeCNPJ,
		Email:        "contact@tbsystems.com.br",
		Status:       models.CustomerStatusActive,
	}
}

func assertStatus(t *testing.T, apiErr interface{ GetStatus() int }, status int) {
	t.Helper()

	if apiErr == nil {
		t.Fatalf("Expected status %d error", status)
	}
	if apiErr.GetStatus() != status {
		t.Fatalf("Expected status %d, got %d", status, apiErr.GetStatus())
	}
}

type fakeCustomerRepository struct {
	createdCustomer models.Customer
	customer        models.Customer
	customers       []models.Customer
	deletedID       uuid.UUID
	documentExists  bool
	emailExists     bool
	findErr         error
	slugExists      bool
	total           int64
}

func (f *fakeCustomerRepository) List(context.Context, commonsmodels.PaginatedParams) ([]models.Customer, int64, error) {
	return f.customers, f.total, nil
}

func (f *fakeCustomerRepository) FindByID(_ context.Context, id uuid.UUID) (models.Customer, error) {
	if f.findErr != nil {
		return models.Customer{}, f.findErr
	}
	if f.customer.ID == uuid.Nil {
		f.customer.ID = id
	}
	return f.customer, nil
}

func (f *fakeCustomerRepository) Create(_ context.Context, customer models.Customer) (models.Customer, error) {
	f.createdCustomer = customer
	return customer, nil
}

func (f *fakeCustomerRepository) Update(_ context.Context, customer models.Customer) (models.Customer, error) {
	return customer, nil
}

func (f *fakeCustomerRepository) Delete(_ context.Context, id uuid.UUID) error {
	f.deletedID = id
	return nil
}

func (f *fakeCustomerRepository) SlugExists(context.Context, string, *uuid.UUID) (bool, error) {
	return f.slugExists, nil
}

func (f *fakeCustomerRepository) DocumentExists(context.Context, string, *uuid.UUID) (bool, error) {
	return f.documentExists, nil
}

func (f *fakeCustomerRepository) EmailExists(context.Context, string, *uuid.UUID) (bool, error) {
	return f.emailExists, nil
}
