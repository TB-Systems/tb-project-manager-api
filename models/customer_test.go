package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestCustomerBeforeCreateGeneratesID(t *testing.T) {
	customer := Customer{}

	if err := customer.BeforeCreate(nil); err != nil {
		t.Fatalf("Expected BeforeCreate to succeed: %v", err)
	}

	if customer.ID == uuid.Nil {
		t.Fatal("Expected customer ID to be generated")
	}
}

func TestCustomerBeforeCreateKeepsExistingID(t *testing.T) {
	id := uuid.New()
	customer := Customer{ID: id}

	if err := customer.BeforeCreate(nil); err != nil {
		t.Fatalf("Expected BeforeCreate to succeed: %v", err)
	}

	if customer.ID != id {
		t.Fatalf("Expected existing customer ID to be kept, got %s", customer.ID)
	}
}
