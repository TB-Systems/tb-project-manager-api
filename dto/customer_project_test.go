package dto

import (
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestCustomerProjectRequestValidate(t *testing.T) {
	t.Run("accepts valid request", func(t *testing.T) {
		request := CustomerProjectRequest{
			ProjectID:    uuid.New(),
			CustomerID:   uuid.New(),
			ProjectValue: 1000,
			MonthlyValue: 100,
			DueDay:       10,
		}

		if errs := request.Validate(); len(errs) != 0 {
			t.Fatalf("Expected no validation errors, got %d", len(errs))
		}
	})

	t.Run("rejects invalid request", func(t *testing.T) {
		request := CustomerProjectRequest{
			ProjectValue: -1,
			MonthlyValue: -1,
			DueDay:       32,
		}

		if errs := request.Validate(); len(errs) != 4 {
			t.Fatalf("Expected 4 validation errors, got %d", len(errs))
		}
	})
}

func TestCustomerProjectBillingStatusFromInvoices(t *testing.T) {
	t.Run("returns monthly ok when setup is fully paid and monthly invoice does not exist yet", func(t *testing.T) {
		status := CustomerProjectBillingStatusFromInvoices(models.CustomerProjectStatusActive, []models.CustomerProjectInvoice{
			{
				Type:   models.CustomerProjectInvoiceTypeSetupFirstHalf,
				Status: models.CustomerProjectInvoiceStatusPaid,
			},
			{
				Type:   models.CustomerProjectInvoiceTypeSetupSecondHalf,
				Status: models.CustomerProjectInvoiceStatusPaid,
			},
		})

		if status != models.CustomerProjectBillingStatusMonthlyOK {
			t.Fatalf("Expected monthly ok billing status, got %d", status)
		}
	})
}
