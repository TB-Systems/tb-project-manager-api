package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceStatus int

const (
	ServiceStatusOnline  ServiceStatus = 1
	ServiceStatusOffline ServiceStatus = 2
)

func (s ServiceStatus) IsValid() bool {
	switch s {
	case ServiceStatusOnline, ServiceStatusOffline:
		return true
	default:
		return false
	}
}

type ServiceCheck struct {
	ID               uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectServiceID uuid.UUID       `gorm:"type:uuid;not null;index" json:"project_service_id"`
	Status           ServiceStatus   `gorm:"not null;check:status IN (1,2)" json:"status"`
	StatusCode       int             `gorm:"not null;default:0" json:"status_code"`
	ResponseTimeMS   int             `gorm:"not null;default:0" json:"response_time_ms"`
	Message          json.RawMessage `gorm:"type:jsonb" json:"message"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	ProjectService   ProjectService  `gorm:"foreignKey:ProjectServiceID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (s *ServiceCheck) BeforeCreate(*gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}

	return nil
}
