package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LogLevel int

const (
	LogLevelInfo     LogLevel = 1
	LogLevelSuccess  LogLevel = 2
	LogLevelWarning  LogLevel = 3
	LogLevelError    LogLevel = 4
	LogLevelCritical LogLevel = 5
)

func (l LogLevel) IsValid() bool {
	switch l {
	case LogLevelInfo,
		LogLevelSuccess,
		LogLevelWarning,
		LogLevelError,
		LogLevelCritical:
		return true
	default:
		return false
	}
}

type ServiceLog struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectServiceID uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_service_id"`
	Level            LogLevel        `gorm:"not null;check:level IN (1,2,3,4,5)" json:"level"`
	Event            string          `gorm:"size:100;not null" json:"event"`
	Message          json.RawMessage `gorm:"type:jsonb;not null" json:"message"`
	Time             time.Time       `gorm:"not null" json:"time"`
	CreatedAt        time.Time       `json:"created_at"`
	ProjectService   ProjectService  `gorm:"foreignKey:ProjectServiceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (s *ServiceLog) BeforeCreate(*gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
