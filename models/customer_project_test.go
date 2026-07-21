package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectPaymentStatusIsValid(t *testing.T) {
	validStatuses := []ProjectPaymentStatus{
		ProjectPaymentStatusFirstHalfPending,
		ProjectPaymentStatusFirstHalfPaid,
		ProjectPaymentStatusSecondHalfPending,
		ProjectPaymentStatusSecondHalfPaid,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Fatalf("Expected status %d to be valid", status)
		}
	}

	if ProjectPaymentStatus(99).IsValid() {
		t.Fatal("Expected unknown project payment status to be invalid")
	}
}

func TestCustomerProjectBeforeCreate(t *testing.T) {
	t.Run("generates ID", func(t *testing.T) {
		customerProject := CustomerProject{}

		if err := customerProject.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if customerProject.ID == uuid.Nil {
			t.Fatal("Expected customer project ID to be generated")
		}
	})

	t.Run("keeps existing ID", func(t *testing.T) {
		id := uuid.New()
		customerProject := CustomerProject{ID: id}

		if err := customerProject.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if customerProject.ID != id {
			t.Fatalf("Expected existing customer project ID to be kept, got %s", customerProject.ID)
		}
	})
}
