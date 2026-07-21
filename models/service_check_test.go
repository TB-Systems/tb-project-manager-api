package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestServiceStatusIsValid(t *testing.T) {
	if !ServiceStatusOnline.IsValid() {
		t.Fatal("Expected online status to be valid")
	}

	if !ServiceStatusOffline.IsValid() {
		t.Fatal("Expected offline status to be valid")
	}

	if ServiceStatus(99).IsValid() {
		t.Fatal("Expected unknown service status to be invalid")
	}
}

func TestServiceCheckBeforeCreate(t *testing.T) {
	t.Run("generates ID", func(t *testing.T) {
		serviceCheck := ServiceCheck{}

		if err := serviceCheck.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if serviceCheck.ID == uuid.Nil {
			t.Fatal("Expected service check ID to be generated")
		}
	})

	t.Run("keeps existing ID", func(t *testing.T) {
		id := uuid.New()
		serviceCheck := ServiceCheck{ID: id}

		if err := serviceCheck.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if serviceCheck.ID != id {
			t.Fatalf("Expected existing service check ID to be kept, got %s", serviceCheck.ID)
		}
	})
}
