package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerDocumentType int

const (
	CustomerDocumentTypeCPF   CustomerDocumentType = 1
	CustomerDocumentTypeCNPJ  CustomerDocumentType = 2
	CustomerDocumentTypeOther CustomerDocumentType = 3
)

func (t CustomerDocumentType) IsValid() bool {
	switch t {
	case CustomerDocumentTypeCPF, CustomerDocumentTypeCNPJ, CustomerDocumentTypeOther:
		return true
	default:
		return false
	}
}

type CustomerStatus int

const (
	CustomerStatusActive             CustomerStatus = 1
	CustomerStatusCanceled           CustomerStatus = 2
	CustomerStatusSuspended          CustomerStatus = 3
	CustomerStatusOnboarding         CustomerStatus = 4
	CustomerStatusLateMonthlyPayment CustomerStatus = 5
)

func (s CustomerStatus) IsValid() bool {
	switch s {
	case CustomerStatusActive,
		CustomerStatusCanceled,
		CustomerStatusSuspended,
		CustomerStatusOnboarding,
		CustomerStatusLateMonthlyPayment:
		return true
	default:
		return false
	}
}

type Customer struct {
	ID           uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	Name         string               `gorm:"size:100;not null" json:"name"`
	Slug         string               `gorm:"size:50;uniqueIndex;not null" json:"slug"`
	Document     string               `gorm:"size:100;uniqueIndex;not null" json:"document"`
	DocumentType CustomerDocumentType `gorm:"not null;check:document_type IN (1,2,3)" json:"document_type"`
	Email        string               `gorm:"size:100;uniqueIndex;not null" json:"email"`
	Phone        string               `gorm:"size:50" json:"phone"`
	Status       CustomerStatus       `gorm:"not null;check:status IN (1,2,3,4,5)" json:"status"`
	URL          string               `gorm:"column:url;size:500" json:"url"`
	CreatedAt    time.Time            `json:"created_at"`
	UpdatedAt    time.Time            `json:"updated_at"`
}

func (c *Customer) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return nil
}
