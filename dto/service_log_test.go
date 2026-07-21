package dto

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestServiceLogRequestValidate(t *testing.T) {
	t.Run("accepts valid request", func(t *testing.T) {
		request := ServiceLogRequest{
			ProjectServiceID: uuid.New(),
			Level:            models.LogLevelInfo,
			Event:            "healthcheck.online",
			Message:          json.RawMessage(`{"ok":true}`),
			Time:             time.Now(),
		}

		if errs := request.Validate(); len(errs) != 0 {
			t.Fatalf("Expected no validation errors, got %d", len(errs))
		}
	})

	t.Run("rejects invalid request", func(t *testing.T) {
		request := ServiceLogRequest{
			Level:   models.LogLevel(99),
			Event:   strings.Repeat("a", 101),
			Message: json.RawMessage(`{"ok":`),
		}

		if errs := request.Validate(); len(errs) != 5 {
			t.Fatalf("Expected 5 validation errors, got %d", len(errs))
		}
	})
}
