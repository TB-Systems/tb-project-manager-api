package dto

import (
	"testing"

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
