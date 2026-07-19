package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

func TestNewServiceLogRepository(t *testing.T) {
	repository := NewServiceLogRepository(nil)
	if repository == nil {
		t.Fatal("Expected service log repository to be initialized")
	}
}

func TestServiceLogRepositoryDryRunQueries(t *testing.T) {
	db := dryRunDB(t)
	repository := NewServiceLogRepository(db)
	ctx := context.Background()
	id := uuid.New()

	t.Run("list builds query", func(t *testing.T) {
		_, _, err := repository.List(ctx, commonsmodels.PaginatedParams{Limit: 10, Offset: 0})
		if err != nil {
			t.Fatalf("Expected dry run list to succeed, got %v", err)
		}
	})

	t.Run("find by id builds query", func(t *testing.T) {
		_, err := repository.FindByID(ctx, id)
		if err != nil {
			t.Fatalf("Expected dry run find to succeed, got %v", err)
		}
	})

	t.Run("create builds query", func(t *testing.T) {
		_, err := repository.Create(ctx, models.ServiceLog{
			ProjectServiceID: uuid.New(),
			Level:            models.LogLevelInfo,
			Event:            "healthcheck.online",
			Message:          json.RawMessage(`{"ok":true}`),
			Time:             time.Now(),
		})
		if err != nil {
			t.Fatalf("Expected dry run create to succeed, got %v", err)
		}
	})
}
