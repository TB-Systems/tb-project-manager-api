package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerProjectStatus int

const (
	CustomerProjectStatusActive CustomerProjectStatus = 1
	CustomerProjectStatusPaused CustomerProjectStatus = 2
	CustomerProjectStatusClosed CustomerProjectStatus = 3
)

func (s CustomerProjectStatus) IsValid() bool {
	switch s {
	case CustomerProjectStatusActive,
		CustomerProjectStatusPaused,
		CustomerProjectStatusClosed:
		return true
	default:
		return false
	}
}

type CustomerProjectBillingStatus int

const (
	CustomerProjectBillingStatusSetupPending       CustomerProjectBillingStatus = 1
	CustomerProjectBillingStatusSetupPartiallyPaid CustomerProjectBillingStatus = 2
	CustomerProjectBillingStatusMonthlyOK          CustomerProjectBillingStatus = 3
	CustomerProjectBillingStatusMonthlyOverdue     CustomerProjectBillingStatus = 4
	CustomerProjectBillingStatusClosed             CustomerProjectBillingStatus = 5
)

func (s CustomerProjectBillingStatus) IsValid() bool {
	switch s {
	case CustomerProjectBillingStatusSetupPending,
		CustomerProjectBillingStatusSetupPartiallyPaid,
		CustomerProjectBillingStatusMonthlyOK,
		CustomerProjectBillingStatusMonthlyOverdue,
		CustomerProjectBillingStatusClosed:
		return true
	default:
		return false
	}
}

type CustomerProject struct {
	ID         uuid.UUID                `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID  uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_customer_project_pair" json:"project_id"`
	CustomerID uuid.UUID                `gorm:"type:uuid;not null;index;uniqueIndex:idx_customer_project_pair" json:"customer_id"`
	Status     CustomerProjectStatus    `gorm:"not null;default:1;check:status IN (1,2,3)" json:"status"`
	StartedAt  time.Time                `gorm:"not null;default:CURRENT_TIMESTAMP" json:"started_at"`
	ClosedAt   *time.Time               `json:"closed_at"`
	CreatedAt  time.Time                `json:"created_at"`
	UpdatedAt  time.Time                `json:"updated_at"`
	Project    Project                  `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Customer   Customer                 `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Terms      []CustomerProjectTerm    `gorm:"foreignKey:CustomerProjectID" json:"-"`
	Invoices   []CustomerProjectInvoice `gorm:"foreignKey:CustomerProjectID" json:"-"`
}

func (c *CustomerProject) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return nil
}
