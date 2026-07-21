package dto

import (
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestCustomerProjectRequestValidate(t *testing.T) {
	t.Run("accepts valid request", func(t *testing.T) {
		request := CustomerProjectRequest{
			ProjectID:            uuid.New(),
			CustomerID:           uuid.New(),
			ProjectValue:         1000,
			MonthlyValue:         100,
			DueDay:               10,
			ProjectPaymentStatus: models.ProjectPaymentStatusFirstHalfPending,
		}

		if errs := request.Validate(); len(errs) != 0 {
			t.Fatalf("Expected no validation errors, got %d", len(errs))
		}
	})

	t.Run("rejects invalid request", func(t *testing.T) {
		request := CustomerProjectRequest{
			ProjectValue:         -1,
			MonthlyValue:         -1,
			DueDay:               32,
			ProjectPaymentStatus: models.ProjectPaymentStatus(99),
		}

		if errs := request.Validate(); len(errs) != 5 {
			t.Fatalf("Expected 5 validation errors, got %d", len(errs))
		}
	})
}
