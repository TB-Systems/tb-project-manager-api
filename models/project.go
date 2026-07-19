package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectType int

const (
	ProjectTypeBackend    ProjectType = 1
	ProjectTypeFrontend   ProjectType = 2
	ProjectTypeAndroid    ProjectType = 3
	ProjectTypeIOS        ProjectType = 4
	ProjectTypeDesktop    ProjectType = 5
	ProjectTypeAutomation ProjectType = 6
	ProjectTypeDatabase   ProjectType = 7
	ProjectTypeOther      ProjectType = 8
)

func (t ProjectType) IsValid() bool {
	switch t {
	case ProjectTypeBackend,
		ProjectTypeFrontend,
		ProjectTypeAndroid,
		ProjectTypeIOS,
		ProjectTypeDesktop,
		ProjectTypeAutomation,
		ProjectTypeDatabase,
		ProjectTypeOther:
		return true
	default:
		return false
	}
}

type ProjectStatus int

const (
	ProjectStatusBacklog    ProjectStatus = 1
	ProjectStatusDiscovery  ProjectStatus = 2
	ProjectStatusDeveloping ProjectStatus = 3
	ProjectStatusStaging    ProjectStatus = 4
	ProjectStatusLive       ProjectStatus = 5
	ProjectStatusPaused     ProjectStatus = 6
	ProjectStatusDown       ProjectStatus = 7
	ProjectStatusArchived   ProjectStatus = 8
)

func (s ProjectStatus) IsValid() bool {
	switch s {
	case ProjectStatusBacklog,
		ProjectStatusDiscovery,
		ProjectStatusDeveloping,
		ProjectStatusStaging,
		ProjectStatusLive,
		ProjectStatusPaused,
		ProjectStatusDown,
		ProjectStatusArchived:
		return true
	default:
		return false
	}
}

type Project struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	Name        string        `gorm:"size:100;not null" json:"name"`
	Description string        `gorm:"size:500" json:"description"`
	Slug        string        `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	RepoURL     string        `gorm:"column:repo_url;size:500" json:"repo_url"`
	Status      ProjectStatus `gorm:"not null;check:status IN (1,2,3,4,5,6,7,8)" json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (p *Project) BeforeCreate(*gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}

	return nil
}
