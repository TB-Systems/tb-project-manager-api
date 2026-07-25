package services

import (
	"testing"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/models"
)

func TestResolveCustomerStatus(t *testing.T) {
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)

	t.Run("onboarding without customer projects", func(t *testing.T) {
		status := ResolveCustomerStatus(nil, now)

		if status != models.CustomerStatusOnboarding {
			t.Fatalf("Expected onboarding status, got %d", status)
		}
	})

	t.Run("active with active project and no overdue invoices", func(t *testing.T) {
		status := ResolveCustomerStatus([]models.CustomerProject{
			{
				Status: models.CustomerProjectStatusActive,
				Invoices: []models.CustomerProjectInvoice{
					{
						Type:   models.CustomerProjectInvoiceTypeSetupFirstHalf,
						Status: models.CustomerProjectInvoiceStatusPaid,
					},
					{
						Type:   models.CustomerProjectInvoiceTypeSetupSecondHalf,
						Status: models.CustomerProjectInvoiceStatusPaid,
					},
					{
						Type:    models.CustomerProjectInvoiceTypeMonthly,
						Status:  models.CustomerProjectInvoiceStatusOpen,
						DueDate: now.AddDate(0, 0, 1),
					},
				},
			},
		}, now)

		if status != models.CustomerStatusActive {
			t.Fatalf("Expected active status, got %d", status)
		}
	})

	t.Run("onboarding with unpaid setup", func(t *testing.T) {
		status := ResolveCustomerStatus([]models.CustomerProject{
			{
				Status: models.CustomerProjectStatusActive,
				Invoices: []models.CustomerProjectInvoice{
					{
						Type:   models.CustomerProjectInvoiceTypeSetupFirstHalf,
						Status: models.CustomerProjectInvoiceStatusPaid,
					},
					{
						Type:   models.CustomerProjectInvoiceTypeSetupSecondHalf,
						Status: models.CustomerProjectInvoiceStatusOpen,
					},
				},
			},
		}, now)

		if status != models.CustomerStatusOnboarding {
			t.Fatalf("Expected onboarding status, got %d", status)
		}
	})

	t.Run("late payment with overdue invoice", func(t *testing.T) {
		status := ResolveCustomerStatus([]models.CustomerProject{
			{
				Status: models.CustomerProjectStatusActive,
				Invoices: []models.CustomerProjectInvoice{
					{
						Type:   models.CustomerProjectInvoiceTypeSetupFirstHalf,
						Status: models.CustomerProjectInvoiceStatusPaid,
					},
					{
						Type:   models.CustomerProjectInvoiceTypeSetupSecondHalf,
						Status: models.CustomerProjectInvoiceStatusPaid,
					},
					{
						Type:    models.CustomerProjectInvoiceTypeMonthly,
						Status:  models.CustomerProjectInvoiceStatusOpen,
						DueDate: now.AddDate(0, 0, -1),
					},
				},
			},
		}, now)

		if status != models.CustomerStatusLateMonthlyPayment {
			t.Fatalf("Expected late payment status, got %d", status)
		}
	})

	t.Run("canceled when every project is closed", func(t *testing.T) {
		status := ResolveCustomerStatus([]models.CustomerProject{
			{Status: models.CustomerProjectStatusClosed},
		}, now)

		if status != models.CustomerStatusCanceled {
			t.Fatalf("Expected canceled status, got %d", status)
		}
	})
}
