package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerProjectInvoiceType int

const (
	CustomerProjectInvoiceTypeSetupFirstHalf  CustomerProjectInvoiceType = 1
	CustomerProjectInvoiceTypeSetupSecondHalf CustomerProjectInvoiceType = 2
	CustomerProjectInvoiceTypeMonthly         CustomerProjectInvoiceType = 3
	CustomerProjectInvoiceTypeDifferential    CustomerProjectInvoiceType = 4
)

func (t CustomerProjectInvoiceType) IsValid() bool {
	switch t {
	case CustomerProjectInvoiceTypeSetupFirstHalf,
		CustomerProjectInvoiceTypeSetupSecondHalf,
		CustomerProjectInvoiceTypeMonthly,
		CustomerProjectInvoiceTypeDifferential:
		return true
	default:
		return false
	}
}

type CustomerProjectInvoiceStatus int

const (
	CustomerProjectInvoiceStatusOpen      CustomerProjectInvoiceStatus = 1
	CustomerProjectInvoiceStatusPaid      CustomerProjectInvoiceStatus = 2
	CustomerProjectInvoiceStatusOverdue   CustomerProjectInvoiceStatus = 3
	CustomerProjectInvoiceStatusCancelled CustomerProjectInvoiceStatus = 4
)

func (s CustomerProjectInvoiceStatus) IsValid() bool {
	switch s {
	case CustomerProjectInvoiceStatusOpen,
		CustomerProjectInvoiceStatusPaid,
		CustomerProjectInvoiceStatusOverdue,
		CustomerProjectInvoiceStatusCancelled:
		return true
	default:
		return false
	}
}

type CustomerProjectInvoice struct {
	ID                uuid.UUID                    `gorm:"type:uuid;primaryKey" json:"id"`
	CustomerProjectID uuid.UUID                    `gorm:"type:uuid;not null;index;uniqueIndex:idx_customer_project_invoice_period" json:"customer_project_id"`
	Type              CustomerProjectInvoiceType   `gorm:"not null;uniqueIndex:idx_customer_project_invoice_period;check:type IN (1,2,3,4)" json:"type"`
	ReferenceMonth    *time.Time                   `gorm:"type:date;uniqueIndex:idx_customer_project_invoice_period" json:"reference_month"`
	Amount            int                          `gorm:"not null;default:0" json:"amount"`
	DueDate           time.Time                    `gorm:"type:date;not null;index" json:"due_date"`
	Status            CustomerProjectInvoiceStatus `gorm:"not null;check:status IN (1,2,3,4)" json:"status"`
	PaidAt            *time.Time                   `json:"paid_at"`
	CreatedAt         time.Time                    `json:"created_at"`
	UpdatedAt         time.Time                    `json:"updated_at"`
	CustomerProject   CustomerProject              `gorm:"foreignKey:CustomerProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (c *CustomerProjectInvoice) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return nil
}
