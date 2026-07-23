package dto

import (
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

type CustomerProjectRequest struct {
	ProjectID    uuid.UUID `json:"project_id"`
	CustomerID   uuid.UUID `json:"customer_id"`
	ProjectValue int       `json:"project_value"`
	MonthlyValue int       `json:"monthly_value"`
	DueDay       int       `json:"due_day"`
}

type CustomerProjectResponse struct {
	ID                   uuid.UUID                   `json:"id"`
	ProjectID            uuid.UUID                   `json:"project_id"`
	CustomerID           uuid.UUID                   `json:"customer_id"`
	ProjectValue         int                         `json:"project_value"`
	MonthlyValue         int                         `json:"monthly_value"`
	DueDay               int                         `json:"due_day"`
	ProjectPaymentStatus models.ProjectPaymentStatus `json:"project_payment_status"`
	LastPayment          *time.Time                  `json:"last_payment"`
	CreatedAt            time.Time                   `json:"created_at"`
	UpdatedAt            time.Time                   `json:"updated_at"`
}

func (request CustomerProjectRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if request.ProjectID == uuid.Nil {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_PROJECT_ID_INVALID"))
	}

	if request.CustomerID == uuid.Nil {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_CUSTOMER_ID_INVALID"))
	}

	if request.ProjectValue < 0 || request.MonthlyValue < 0 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_VALUE_INVALID"))
	}

	if request.DueDay < 1 || request.DueDay > 31 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_DUE_DAY_INVALID"))
	}

	return errs
}

func CustomerProjectResponseFromModel(customerProject models.CustomerProject) CustomerProjectResponse {
	return CustomerProjectResponse{
		ID:                   customerProject.ID,
		ProjectID:            customerProject.ProjectID,
		CustomerID:           customerProject.CustomerID,
		ProjectValue:         customerProject.ProjectValue,
		MonthlyValue:         customerProject.MonthlyValue,
		DueDay:               customerProject.DueDay,
		ProjectPaymentStatus: customerProject.ProjectPaymentStatus,
		LastPayment:          customerProject.LastPayment,
		CreatedAt:            customerProject.CreatedAt,
		UpdatedAt:            customerProject.UpdatedAt,
	}
}
