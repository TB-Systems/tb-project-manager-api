package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestLogLevelIsValid(t *testing.T) {
	validLevels := []LogLevel{
		LogLevelInfo,
		LogLevelSuccess,
		LogLevelWarning,
		LogLevelError,
		LogLevelCritical,
	}

	for _, level := range validLevels {
		if !level.IsValid() {
			t.Fatalf("Expected log level %d to be valid", level)
		}
	}

	if LogLevel(99).IsValid() {
		t.Fatal("Expected unknown log level to be invalid")
	}
}

func TestServiceLogBeforeCreate(t *testing.T) {
	t.Run("generates ID", func(t *testing.T) {
		serviceLog := ServiceLog{}

		if err := serviceLog.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if serviceLog.ID == uuid.Nil {
			t.Fatal("Expected service log ID to be generated")
		}
	})

	t.Run("keeps existing ID", func(t *testing.T) {
		id := uuid.New()
		serviceLog := ServiceLog{ID: id}

		if err := serviceLog.BeforeCreate(nil); err != nil {
			t.Fatalf("Expected BeforeCreate to succeed: %v", err)
		}

		if serviceLog.ID != id {
			t.Fatalf("Expected existing service log ID to be kept, got %s", serviceLog.ID)
		}
	})
}
