package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProjectPaymentStatus int

const (
	ProjectPaymentStatusFirstHalfPending  ProjectPaymentStatus = 1
	ProjectPaymentStatusFirstHalfPaid     ProjectPaymentStatus = 2
	ProjectPaymentStatusSecondHalfPending ProjectPaymentStatus = 3
	ProjectPaymentStatusSecondHalfPaid    ProjectPaymentStatus = 4
	PaymentOnDay                          ProjectPaymentStatus = 5
	PaymentPeding                         ProjectPaymentStatus = 6
)

func (s ProjectPaymentStatus) IsValid() bool {
	switch s {
	case ProjectPaymentStatusFirstHalfPending,
		ProjectPaymentStatusFirstHalfPaid,
		ProjectPaymentStatusSecondHalfPending,
		ProjectPaymentStatusSecondHalfPaid:
		return true
	default:
		return false
	}
}

type CustomerProject struct {
	ID                   uuid.UUID            `gorm:"type:uuid;primaryKey" json:"id"`
	ProjectID            uuid.UUID            `gorm:"type:uuid;not null;uniqueIndex;uniqueIndex:idx_customer_project_pair" json:"project_id"`
	CustomerID           uuid.UUID            `gorm:"type:uuid;not null;index;uniqueIndex:idx_customer_project_pair" json:"customer_id"`
	ProjectValue         int                  `gorm:"not null;default:0" json:"project_value"`
	MonthlyValue         int                  `gorm:"not null;default:0" json:"monthly_value"`
	DueDay               int                  `gorm:"not null" json:"due_day"`
	ProjectPaymentStatus ProjectPaymentStatus `gorm:"not null;check:project_payment_status IN (1,2,3,4)" json:"project_payment_status"`
	LastPayment          *time.Time           `json:"last_payment"`
	CreatedAt            time.Time            `json:"created_at"`
	UpdatedAt            time.Time            `json:"updated_at"`
	Project              Project              `gorm:"foreignKey:ProjectID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
	Customer             Customer             `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT" json:"-"`
}

func (c *CustomerProject) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return nil
}
