package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestDailyBillingJobRunCreatesMonthlyInvoices(t *testing.T) {
	customerProjectID := uuid.New()
	customerID := uuid.New()
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	repository := &fakeDailyBillingInvoiceRepository{
		candidates: []models.CustomerProject{
			{
				ID:         customerProjectID,
				CustomerID: customerID,
				Status:     models.CustomerProjectStatusActive,
				Project:    models.Project{Status: models.ProjectStatusLive},
				Terms: []models.CustomerProjectTerm{
					{Active: true, MonthlyValue: 15000, DueDay: 31},
				},
				Invoices: []models.CustomerProjectInvoice{
					{Type: models.CustomerProjectInvoiceTypeSetupFirstHalf, Status: models.CustomerProjectInvoiceStatusPaid},
					{Type: models.CustomerProjectInvoiceTypeSetupSecondHalf, Status: models.CustomerProjectInvoiceStatusPaid},
				},
			},
		},
		customerIDs: []uuid.UUID{customerID},
	}
	customerStatus := &fakeDailyBillingCustomerStatus{}
	billingSync := &fakeDailyBillingSync{}
	job := NewDailyBillingJob(repository, billingSync, customerStatus)
	job.now = func() time.Time { return now }

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Expected daily billing job to run: %v", err)
	}

	if len(repository.createdInvoices) != 1 {
		t.Fatalf("Expected 1 monthly invoice, got %d", len(repository.createdInvoices))
	}

	invoice := repository.createdInvoices[0]
	if invoice.Type != models.CustomerProjectInvoiceTypeMonthly {
		t.Fatalf("Expected monthly invoice type, got %d", invoice.Type)
	}
	if invoice.ReferenceMonth == nil || invoice.ReferenceMonth.Day() != 1 || invoice.ReferenceMonth.Month() != time.July {
		t.Fatalf("Expected July reference month, got %v", invoice.ReferenceMonth)
	}
	if invoice.DueDate.Day() != 31 {
		t.Fatalf("Expected due day clamped to July 31, got %s", invoice.DueDate)
	}
	if len(customerStatus.syncedCustomerIDs) != 1 || customerStatus.syncedCustomerIDs[0] != customerID {
		t.Fatalf("Expected customer %s status sync, got %#v", customerID, customerStatus.syncedCustomerIDs)
	}
	if !billingSync.synced {
		t.Fatal("Expected overdue sync to run")
	}
}

func TestDailyBillingJobSkipsUnpaidSetup(t *testing.T) {
	customerProjectID := uuid.New()
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	repository := &fakeDailyBillingInvoiceRepository{
		candidates: []models.CustomerProject{
			{
				ID:      customerProjectID,
				Status:  models.CustomerProjectStatusActive,
				Project: models.Project{Status: models.ProjectStatusLive},
				Terms: []models.CustomerProjectTerm{
					{Active: true, MonthlyValue: 15000, DueDay: 10},
				},
				Invoices: []models.CustomerProjectInvoice{
					{Type: models.CustomerProjectInvoiceTypeSetupFirstHalf, Status: models.CustomerProjectInvoiceStatusPaid},
					{Type: models.CustomerProjectInvoiceTypeSetupSecondHalf, Status: models.CustomerProjectInvoiceStatusOpen},
				},
			},
		},
	}
	job := NewDailyBillingJob(repository, &fakeDailyBillingSync{}, &fakeDailyBillingCustomerStatus{})
	job.now = func() time.Time { return now }

	if err := job.Run(context.Background()); err != nil {
		t.Fatalf("Expected daily billing job to run: %v", err)
	}

	if len(repository.createdInvoices) != 0 {
		t.Fatalf("Expected no monthly invoices, got %d", len(repository.createdInvoices))
	}
}

type fakeDailyBillingInvoiceRepository struct {
	candidates      []models.CustomerProject
	createdInvoices []models.CustomerProjectInvoice
	customerIDs     []uuid.UUID
}

func (f *fakeDailyBillingInvoiceRepository) FindByID(context.Context, uuid.UUID) (models.CustomerProjectInvoice, error) {
	return models.CustomerProjectInvoice{}, nil
}

func (f *fakeDailyBillingInvoiceRepository) Create(context.Context, models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error) {
	return models.CustomerProjectInvoice{}, nil
}

func (f *fakeDailyBillingInvoiceRepository) Pay(context.Context, uuid.UUID, time.Time) (models.CustomerProjectInvoice, error) {
	return models.CustomerProjectInvoice{}, nil
}

func (f *fakeDailyBillingInvoiceRepository) Update(context.Context, models.CustomerProjectInvoice) (models.CustomerProjectInvoice, error) {
	return models.CustomerProjectInvoice{}, nil
}

func (f *fakeDailyBillingInvoiceRepository) Unpay(context.Context, uuid.UUID, models.CustomerProjectInvoiceStatus) (models.CustomerProjectInvoice, error) {
	return models.CustomerProjectInvoice{}, nil
}

func (f *fakeDailyBillingInvoiceRepository) MarkOverdue(context.Context, time.Time) ([]uuid.UUID, error) {
	return nil, nil
}

func (f *fakeDailyBillingInvoiceRepository) ListMonthlyInvoiceCandidates(context.Context) ([]models.CustomerProject, error) {
	return f.candidates, nil
}

func (f *fakeDailyBillingInvoiceRepository) CreateMonthlyInvoices(_ context.Context, invoices []models.CustomerProjectInvoice) ([]uuid.UUID, error) {
	f.createdInvoices = invoices
	return f.customerIDs, nil
}

type fakeDailyBillingSync struct {
	synced bool
}

func (f *fakeDailyBillingSync) SyncOverdue(context.Context) errors.ApiError {
	f.synced = true
	return nil
}

type fakeDailyBillingCustomerStatus struct {
	syncedCustomerIDs []uuid.UUID
}

func (f *fakeDailyBillingCustomerStatus) Sync(_ context.Context, customerID uuid.UUID) errors.ApiError {
	f.syncedCustomerIDs = append(f.syncedCustomerIDs, customerID)
	return nil
}

func (f *fakeDailyBillingCustomerStatus) SyncMany(_ context.Context, customerIDs []uuid.UUID) errors.ApiError {
	f.syncedCustomerIDs = append(f.syncedCustomerIDs, customerIDs...)
	return nil
}
