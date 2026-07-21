package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectService struct {
	ID             uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID      uuid.UUID     `gorm:"type:uuid;not null;index" json:"project_id"`
	Name           string        `gorm:"size:100;not null" json:"name"`
	Type           ProjectType   `gorm:"not null;check:type IN (1,2,3,4,5,6,7,8)" json:"type"`
	URL            string        `gorm:"column:url;size:500" json:"url"`
	RepoURL        string        `gorm:"column:repo_url;size:500" json:"repo_url"`
	Status         ProjectStatus `gorm:"not null;check:status IN (1,2,3,4,5,6,7,8)" json:"status"`
	HealthCheckURL string        `gorm:"column:health_check_url;size:500" json:"health_check_url"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Project        Project       `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (p *ProjectService) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}
