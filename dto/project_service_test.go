package dto

import (
	"strings"
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestProjectServiceRequestValidate(t *testing.T) {
	t.Run("accepts valid request", func(t *testing.T) {
		request := ProjectServiceRequest{
			ProjectID:      uuid.New(),
			Name:           "Project Manager API",
			Type:           models.ProjectTypeBackend,
			URL:            "https://api.example.com",
			RepoURL:        "https://github.com/TB-Systems/tb-project-manager-api",
			Status:         models.ProjectStatusLive,
			HealthCheckURL: "https://api.example.com/health",
		}

		if errs := request.Validate(); len(errs) != 0 {
			t.Fatalf("Expected no validation errors, got %d", len(errs))
		}
	})

	t.Run("accepts valid create request without status", func(t *testing.T) {
		request := ProjectServiceCreateRequest{
			ProjectID:      uuid.New(),
			Name:           "Project Manager API",
			Type:           models.ProjectTypeBackend,
			URL:            "https://api.example.com",
			RepoURL:        "https://github.com/TB-Systems/tb-project-manager-api",
			HealthCheckURL: "https://api.example.com/health",
		}

		if errs := request.Validate(); len(errs) != 0 {
			t.Fatalf("Expected no validation errors, got %d", len(errs))
		}
	})

	t.Run("rejects invalid request", func(t *testing.T) {
		request := ProjectServiceRequest{
			Name:           strings.Repeat("a", 101),
			Type:           models.ProjectType(99),
			URL:            strings.Repeat("a", 501),
			RepoURL:        strings.Repeat("a", 501),
			Status:         models.ProjectStatus(99),
			HealthCheckURL: strings.Repeat("a", 501),
		}

		if errs := request.Validate(); len(errs) != 7 {
			t.Fatalf("Expected 7 validation errors, got %d", len(errs))
		}
	})

	t.Run("rejects invalid create request without status validation", func(t *testing.T) {
		request := ProjectServiceCreateRequest{
			Name:           strings.Repeat("a", 101),
			Type:           models.ProjectType(99),
			URL:            strings.Repeat("a", 501),
			RepoURL:        strings.Repeat("a", 501),
			HealthCheckURL: strings.Repeat("a", 501),
		}

		if errs := request.Validate(); len(errs) != 6 {
			t.Fatalf("Expected 6 validation errors, got %d", len(errs))
		}
	})
}
