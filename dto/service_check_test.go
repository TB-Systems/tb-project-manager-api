package dto

import (
	"encoding/json"
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestServiceCheckRequestValidate(t *testing.T) {
	t.Run("accepts valid request", func(t *testing.T) {
		request := ServiceCheckRequest{
			ProjectServiceID: uuid.New(),
			Status:           models.ServiceStatusOnline,
			StatusCode:       200,
			ResponseTimeMS:   120,
			Message:          json.RawMessage(`{"ok":true}`),
		}

		if errs := request.Validate(); len(errs) != 0 {
			t.Fatalf("Expected no validation errors, got %d", len(errs))
		}
	})

	t.Run("rejects invalid request", func(t *testing.T) {
		request := ServiceCheckRequest{
			Status:         models.ServiceStatus(99),
			StatusCode:     -1,
			ResponseTimeMS: -1,
			Message:        json.RawMessage(`{"ok":`),
		}

		if errs := request.Validate(); len(errs) != 5 {
			t.Fatalf("Expected 5 validation errors, got %d", len(errs))
		}
	})
}
