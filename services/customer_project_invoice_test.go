package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func TestCustomerProjectInvoiceServicePay(t *testing.T) {
	t.Run("marks invoice as paid", func(t *testing.T) {
		id := uuid.New()
		now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
		repository := &fakeCustomerProjectInvoiceRepository{
			invoice: models.CustomerProjectInvoice{
				ID:     id,
				Status: models.CustomerProjectInvoiceStatusOpen,
			},
		}
		service := customerProjectInvoice{repository: repository, now: func() time.Time { return now }}

		response, apiErr := service.Pay(context.Background(), id.String(), dto.CustomerProjectInvoicePayRequest{})

		if apiErr != nil {
			t.Fatalf("Expected invoice payment, got status %d", apiErr.GetStatus())
		}
		if response.Status != models.CustomerProjectInvoiceStatusPaid {
			t.Fatalf("Expected paid status, got %d", response.Status)
		}
		if response.PaidAt == nil || !response.PaidAt.Equal(now) {
			t.Fatalf("Expected paid_at %s, got %v", now, response.PaidAt)
		}
		if !repository.markOverdueCalled {
			t.Fatal("Expected overdue sync before payment")
		}
	})

	t.Run("uses requested paid_at", func(t *testing.T) {
		id := uuid.New()
		paidAt := time.Date(2026, 7, 20, 9, 0, 0, 0, time.UTC)
		repository := &fakeCustomerProjectInvoiceRepository{
			invoice: models.CustomerProjectInvoice{
				ID:     id,
				Status: models.CustomerProjectInvoiceStatusOverdue,
			},
		}
		service := customerProjectInvoice{repository: repository, now: time.Now}

		response, apiErr := service.Pay(context.Background(), id.String(), dto.CustomerProjectInvoicePayRequest{PaidAt: &paidAt})

		if apiErr != nil {
			t.Fatalf("Expected invoice payment, got status %d", apiErr.GetStatus())
		}
		if response.PaidAt == nil || !response.PaidAt.Equal(paidAt) {
			t.Fatalf("Expected requested paid_at %s, got %v", paidAt, response.PaidAt)
		}
	})

	t.Run("rejects cancelled invoice", func(t *testing.T) {
		id := uuid.New()
		repository := &fakeCustomerProjectInvoiceRepository{
			invoice: models.CustomerProjectInvoice{
				ID:     id,
				Status: models.CustomerProjectInvoiceStatusCancelled,
			},
		}
		service := customerProjectInvoice{repository: repository, now: time.Now}

		_, apiErr := service.Pay(context.Background(), id.String(), dto.CustomerProjectInvoicePayRequest{})

		if apiErr == nil {
			t.Fatal("Expected cancelled invoice error")
		}
		if apiErr.GetStatus() != http.StatusBadRequest {
			t.Fatalf("Expected status %d, got %d", http.StatusBadRequest, apiErr.GetStatus())
		}
	})

	t.Run("returns not found", func(t *testing.T) {
		repository := &fakeCustomerProjectInvoiceRepository{findErr: gorm.ErrRecordNotFound}
		service := customerProjectInvoice{repository: repository, now: time.Now}

		_, apiErr := service.Pay(context.Background(), uuid.NewString(), dto.CustomerProjectInvoicePayRequest{})

		if apiErr == nil {
			t.Fatal("Expected not found error")
		}
		if apiErr.GetStatus() != http.StatusNotFound {
			t.Fatalf("Expected status %d, got %d", http.StatusNotFound, apiErr.GetStatus())
		}
	})
}

func TestCustomerProjectInvoiceServiceSyncOverdue(t *testing.T) {
	now := time.Date(2026, 7, 25, 15, 0, 0, 0, time.UTC)
	repository := &fakeCustomerProjectInvoiceRepository{}
	service := customerProjectInvoice{repository: repository, now: func() time.Time { return now }}

	apiErr := service.SyncOverdue(context.Background())

	if apiErr != nil {
		t.Fatalf("Expected overdue sync, got status %d", apiErr.GetStatus())
	}
	expectedToday := time.Date(2026, 7, 25, 0, 0, 0, 0, time.UTC)
	if !repository.markOverdueDate.Equal(expectedToday) {
		t.Fatalf("Expected overdue sync date %s, got %s", expectedToday, repository.markOverdueDate)
	}
}

type fakeCustomerProjectInvoiceRepository struct {
	invoice           models.CustomerProjectInvoice
	findErr           error
	payErr            error
	markOverdueErr    error
	markOverdueCalled bool
	markOverdueDate   time.Time
}

func (f *fakeCustomerProjectInvoiceRepository) FindByID(_ context.Context, id uuid.UUID) (models.CustomerProjectInvoice, error) {
	if f.findErr != nil {
		return models.CustomerProjectInvoice{}, f.findErr
	}
	if f.invoice.ID == uuid.Nil {
		f.invoice.ID = id
	}
	return f.invoice, nil
}

func (f *fakeCustomerProjectInvoiceRepository) Create(_ context.Context, invoice models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error) {
	f.invoice = invoice
	return f.invoice, nil
}

func (f *fakeCustomerProjectInvoiceRepository) Pay(_ context.Context, id uuid.UUID, paidAt time.Time) (models.CustomerProjectInvoice, error) {
	if f.payErr != nil {
		return models.CustomerProjectInvoice{}, f.payErr
	}

	f.invoice.ID = id
	f.invoice.Status = models.CustomerProjectInvoiceStatusPaid
	f.invoice.PaidAt = &paidAt
	return f.invoice, nil
}

func (f *fakeCustomerProjectInvoiceRepository) Update(_ context.Context, invoice models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error) {
	f.invoice = invoice
	return f.invoice, nil
}

func (f *fakeCustomerProjectInvoiceRepository) Unpay(_ context.Context, id uuid.UUID, status models.CustomerProjectInvoiceStatus) (models.CustomerProjectInvoice, error) {
	f.invoice.ID = id
	f.invoice.Status = status
	f.invoice.PaidAt = nil
	return f.invoice, nil
}

func (f *fakeCustomerProjectInvoiceRepository) MarkOverdue(_ context.Context, today time.Time) ([]uuid.UUID, error) {
	f.markOverdueCalled = true
	f.markOverdueDate = today
	return nil, f.markOverdueErr
}

func (f *fakeCustomerProjectInvoiceRepository) ListMonthlyInvoiceCandidates(context.Context) ([]models.CustomerProject, error) {
	return nil, nil
}

func (f *fakeCustomerProjectInvoiceRepository) CreateMonthlyInvoices(context.Context, []models.CustomerProjectInvoice) ([]uuid.UUID, error) {
	return nil, nil
}
