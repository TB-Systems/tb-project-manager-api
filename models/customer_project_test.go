package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestCustomerProjectStatusIsValid(t *testing.T) {
	validStatuses := []CustomerProjectStatus{
		CustomerProjectStatusActive,
		CustomerProjectStatusPaused,
		CustomerProjectStatusClosed,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Fatalf("Expected status %d to be valid", status)
		}
	}

	if CustomerProjectStatus(99).IsValid() {
		t.Fatal("Expected unknown customer project status to be invalid")
	}
}

func TestCustomerProjectBillingStatusIsValid(t *testing.T) {
	validStatuses := []CustomerProjectBillingStatus{
		CustomerProjectBillingStatusSetupPending,
		CustomerProjectBillingStatusSetupPartiallyPaid,
		CustomerProjectBillingStatusMonthlyOK,
		CustomerProjectBillingStatusMonthlyOverdue,
		CustomerProjectBillingStatusClosed,
	}

	for _, status := range validStatuses {
		if !status.IsValid() {
			t.Fatalf("Expected status %d to be valid", status)
		}
	}

	if CustomerProjectBillingStatus(99).IsValid() {
		t.Fatal("Expected unknown billing status to be invalid")
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
