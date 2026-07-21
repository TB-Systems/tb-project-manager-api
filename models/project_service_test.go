package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestProjectTypeIsValidIncludesServiceTypes(t *testing.T) {
	validTypes := []ProjectType{
		ProjectTypeBackend,
		ProjectTypeFrontend,
		ProjectTypeAndroid,
		ProjectTypeIOS,
		ProjectTypeDesktop,
		ProjectTypeAutomation,
		ProjectTypeDatabase,
		ProjectTypeOther,
	}

	for _, projectType := range validTypes {
		if !projectType.IsValid() {
			t.Fatalf("Expected project type %d to be valid", projectType)
		}
	}

	if ProjectType(99).IsValid() {
		t.Fatal("Expected unknown project type to be invalid")
	}
}

func TestProjectServiceBeforeCreate(t *testing.T) {
	t.Run("generates ID", func(t *testing.T) {
		projectService := ProjectService{}

		if err := projectService.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if projectService.ID == uuid.Nil {
			t.Fatal("Expected project service ID to be generated")
		}
	})

	t.Run("keeps existing ID", func(t *testing.T) {
		id := uuid.New()
		projectService := ProjectService{ID: id}

		if err := projectService.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if projectService.ID != id {
			t.Fatalf("Expected existing project service ID to be kept, got %s", projectService.ID)
		}
	})
}
