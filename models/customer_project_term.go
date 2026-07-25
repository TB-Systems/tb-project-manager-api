package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerProjectTerm struct {
	ID                uuid.UUID       `gorm:"type:uuid;primaryKey" json:"id"`
	CustomerProjectID uuid.UUID       `gorm:"type:uuid;not null;index" json:"customer_project_id"`
	SetupValue        int             `gorm:"not null;default:0" json:"setup_value"`
	MonthlyValue      int             `gorm:"not null;default:0" json:"monthly_value"`
	DueDay            int             `gorm:"not null" json:"due_day"`
	StartsAt          time.Time       `gorm:"not null" json:"starts_at"`
	EndsAt            *time.Time      `json:"ends_at"`
	Active            bool            `gorm:"not null;default:true" json:"active"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	CustomerProject   CustomerProject `gorm:"foreignKey:CustomerProjectID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE" json:"-"`
}

func (c *CustomerProjectTerm) BeforeCreate(*gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}

	return nil
}
