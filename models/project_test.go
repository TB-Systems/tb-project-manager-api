package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectBeforeCreateGeneratesID(t *testing.T) {
	project := Project{}

	if err := project.BeforeCreate(nil); err != nil {
		t.Fatalf("Expected BeforeCreate to succeed: %v", err)
	}

	if project.ID == uuid.Nil {
		t.Fatal("Expected project ID to be generated")
	}
}

func TestProjectBeforeCreateKeepsExistingID(t *testing.T) {
	id := uuid.New()
	project := Project{ID: id}

	if err := project.BeforeCreate(nil); err != nil {
		t.Fatalf("Expected BeforeCreate to succeed: %v", err)
	}

	if project.ID != id {
		t.Fatalf("Expected existing project ID to be kept, got %s", project.ID)
	}
}
