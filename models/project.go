package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectType int

const (
	ProjectTypeBackend  ProjectType = 1
	ProjectTypeFrontend ProjectType = 2
	ProjectTypeMobile   ProjectType = 3
)

type ProjectStatus int

const (
	ProjectStatusBacklog    ProjectStatus = 1
	ProjectStatusDeveloping ProjectStatus = 2
	ProjectStatusStaging    ProjectStatus = 3
	ProjectStatusLive       ProjectStatus = 4
)

type Project struct {
	ID                    uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Name                  string        `gorm:"size:100;not null" json:"name"`
	Slug                  string        `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	Type                  ProjectType   `gorm:"not null;check:type IN (1,2,3)" json:"type"`
	BaseURL               string        `gorm:"column:base_url;size:500" json:"base_url"`
	RepoURL               string        `gorm:"column:repo_url;size:500" json:"repo_url"`
	SharedValue           int           `gorm:"not null;default:0" json:"shared_value"`
	DedicatedValue        int           `gorm:"not null;default:0" json:"dedicated_value"`
	SupportSharedValue    int           `gorm:"not null;default:0" json:"support_shared_value"`
	SupportDedicatedValue int           `gorm:"not null;default:0" json:"support_dedicated_value"`
	Status                ProjectStatus `gorm:"not null;check:status IN (1,2,3,4)" json:"status"`
	CreatedAt             time.Time     `json:"created_at"`
	UpdatedAt             time.Time     `json:"updated_at"`
}

func (p *Project) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}
