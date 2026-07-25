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
	ID            uuid.UUID                           `json:"id"`
	ProjectID     uuid.UUID                           `json:"project_id"`
	CustomerID    uuid.UUID                           `json:"customer_id"`
	Status        models.CustomerProjectStatus        `json:"status"`
	BillingStatus models.CustomerProjectBillingStatus `json:"billing_status"`
	ProjectValue  int                                 `json:"project_value"`
	MonthlyValue  int                                 `json:"monthly_value"`
	DueDay        int                                 `json:"due_day"`
	Terms         []CustomerProjectTermResponse       `json:"terms"`
	Invoices      []CustomerProjectInvoiceResponse    `json:"invoices"`
	StartedAt     time.Time                           `json:"started_at"`
	ClosedAt      *time.Time                          `json:"closed_at"`
	CreatedAt     time.Time                           `json:"created_at"`
	UpdatedAt     time.Time                           `json:"updated_at"`
}

type CustomerProjectTermResponse struct {
	ID                uuid.UUID  `json:"id"`
	CustomerProjectID uuid.UUID  `json:"customer_project_id"`
	SetupValue        int        `json:"setup_value"`
	MonthlyValue      int        `json:"monthly_value"`
	DueDay            int        `json:"due_day"`
	StartsAt          time.Time  `json:"starts_at"`
	EndsAt            *time.Time `json:"ends_at"`
	Active            bool       `json:"active"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

type CustomerProjectInvoiceResponse struct {
	ID                uuid.UUID                           `json:"id"`
	CustomerProjectID uuid.UUID                           `json:"customer_project_id"`
	Type              models.CustomerProjectInvoiceType   `json:"type"`
	ReferenceMonth    *time.Time                          `json:"reference_month"`
	Amount            int                                 `json:"amount"`
	DueDate           time.Time                           `json:"due_date"`
	Status            models.CustomerProjectInvoiceStatus `json:"status"`
	PaidAt            *time.Time                          `json:"paid_at"`
	CreatedAt         time.Time                           `json:"created_at"`
	UpdatedAt         time.Time                           `json:"updated_at"`
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
	terms := customerProjectTermsFromModel(customerProject.Terms)
	invoices := customerProjectInvoicesFromModel(customerProject.Invoices)
	projectValue, monthlyValue, dueDay := currentCommercialTerms(customerProject.Terms)

	return CustomerProjectResponse{
		ID:            customerProject.ID,
		ProjectID:     customerProject.ProjectID,
		CustomerID:    customerProject.CustomerID,
		Status:        customerProject.Status,
		BillingStatus: CustomerProjectBillingStatusFromInvoices(customerProject.Status, customerProject.Invoices),
		ProjectValue:  projectValue,
		MonthlyValue:  monthlyValue,
		DueDay:        dueDay,
		Terms:         terms,
		Invoices:      invoices,
		StartedAt:     customerProject.StartedAt,
		ClosedAt:      customerProject.ClosedAt,
		CreatedAt:     customerProject.CreatedAt,
		UpdatedAt:     customerProject.UpdatedAt,
	}
}

func customerProjectTermsFromModel(terms []models.CustomerProjectTerm) []CustomerProjectTermResponse {
	items := make([]CustomerProjectTermResponse, 0, len(terms))
	for _, term := range terms {
		items = append(items, CustomerProjectTermResponse{
			ID:                term.ID,
			CustomerProjectID: term.CustomerProjectID,
			SetupValue:        term.SetupValue,
			MonthlyValue:      term.MonthlyValue,
			DueDay:            term.DueDay,
			StartsAt:          term.StartsAt,
			EndsAt:            term.EndsAt,
			Active:            term.Active,
			CreatedAt:         term.CreatedAt,
			UpdatedAt:         term.UpdatedAt,
		})
	}

	return items
}

func customerProjectInvoicesFromModel(invoices []models.CustomerProjectInvoice) []CustomerProjectInvoiceResponse {
	items := make([]CustomerProjectInvoiceResponse, 0, len(invoices))
	for _, invoice := range invoices {
		items = append(items, CustomerProjectInvoiceResponseFromModel(invoice))
	}

	return items
}

func CustomerProjectInvoiceResponseFromModel(invoice models.CustomerProjectInvoice) CustomerProjectInvoiceResponse {
	return CustomerProjectInvoiceResponse{
		ID:                invoice.ID,
		CustomerProjectID: invoice.CustomerProjectID,
		Type:              invoice.Type,
		ReferenceMonth:    invoice.ReferenceMonth,
		Amount:            invoice.Amount,
		DueDate:           invoice.DueDate,
		Status:            invoice.Status,
		PaidAt:            invoice.PaidAt,
		CreatedAt:         invoice.CreatedAt,
		UpdatedAt:         invoice.UpdatedAt,
	}
}

func currentCommercialTerms(terms []models.CustomerProjectTerm) (int, int, int) {
	for _, term := range terms {
		if term.Active {
			return term.SetupValue, term.MonthlyValue, term.DueDay
		}
	}

	if len(terms) == 0 {
		return 0, 0, 0
	}

	term := terms[0]
	return term.SetupValue, term.MonthlyValue, term.DueDay
}

func CustomerProjectBillingStatusFromInvoices(status models.CustomerProjectStatus, invoices []models.CustomerProjectInvoice) models.CustomerProjectBillingStatus {
	if status == models.CustomerProjectStatusClosed {
		return models.CustomerProjectBillingStatusClosed
	}

	firstHalfPaid := false
	secondHalfPaid := false
	hasSetupFirstHalf := false
	hasSetupSecondHalf := false
	hasMonthly := false
	hasMonthlyOverdue := false
	today := beginningOfUTCDay(time.Now().UTC())

	for _, invoice := range invoices {
		invoiceOverdue := invoice.Status == models.CustomerProjectInvoiceStatusOverdue ||
			(invoice.Status == models.CustomerProjectInvoiceStatusOpen && invoice.DueDate.Before(today))

		switch invoice.Type {
		case models.CustomerProjectInvoiceTypeSetupFirstHalf:
			hasSetupFirstHalf = true
			firstHalfPaid = invoice.Status == models.CustomerProjectInvoiceStatusPaid
		case models.CustomerProjectInvoiceTypeSetupSecondHalf:
			hasSetupSecondHalf = true
			secondHalfPaid = invoice.Status == models.CustomerProjectInvoiceStatusPaid
		case models.CustomerProjectInvoiceTypeMonthly:
			hasMonthly = true
			if invoiceOverdue {
				hasMonthlyOverdue = true
			}
		}
	}

	if hasSetupFirstHalf && !firstHalfPaid {
		return models.CustomerProjectBillingStatusSetupPending
	}
	if hasSetupSecondHalf && !secondHalfPaid {
		return models.CustomerProjectBillingStatusSetupPartiallyPaid
	}
	if hasMonthlyOverdue {
		return models.CustomerProjectBillingStatusMonthlyOverdue
	}
	if hasMonthly {
		return models.CustomerProjectBillingStatusMonthlyOK
	}

	return models.CustomerProjectBillingStatusMonthlyOK
}

func beginningOfUTCDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
