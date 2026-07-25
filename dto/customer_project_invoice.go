package dto

import (
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/models"
)

type CustomerProjectInvoicePayRequest struct {
	PaidAt *time.Time `json:"paid_at"`
}

type CustomerProjectInvoiceRequest struct {
	Type           models.CustomerProjectInvoiceType   `json:"type"`
	ReferenceMonth *time.Time                          `json:"reference_month"`
	Amount         int                                 `json:"amount"`
	DueDate        time.Time                           `json:"due_date"`
	Status         models.CustomerProjectInvoiceStatus `json:"status"`
	PaidAt         *time.Time                          `json:"paid_at"`
}

func (request CustomerProjectInvoicePayRequest) Validate() []errors.ApiErrorItem {
	return make([]errors.ApiErrorItem, 0)
}

func (request CustomerProjectInvoiceRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if !request.Type.IsValid() {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_INVOICE_TYPE_INVALID"))
	}

	if request.Amount < 0 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_INVOICE_AMOUNT_INVALID"))
	}

	if request.DueDate.IsZero() {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_INVOICE_DUE_DATE_INVALID"))
	}

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_INVOICE_STATUS_INVALID"))
	}

	if request.Status == models.CustomerProjectInvoiceStatusPaid && request.PaidAt == nil {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PROJECT_INVOICE_PAID_AT_REQUIRED"))
	}

	return errs
}
