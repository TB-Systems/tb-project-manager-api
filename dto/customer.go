package dto

import (
	"strings"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/google/uuid"
)

type CustomerRequest struct {
	Name         string                      `json:"name"`
	Slug         string                      `json:"slug"`
	Document     string                      `json:"document"`
	DocumentType models.CustomerDocumentType `json:"document_type"`
	Email        string                      `json:"email"`
	Phone        string                      `json:"phone"`
	Status       models.CustomerStatus       `json:"status"`
	URL          string                      `json:"url"`
}

type CustomerResponse struct {
	ID           uuid.UUID                   `json:"id"`
	Name         string                      `json:"name"`
	Slug         string                      `json:"slug"`
	Document     string                      `json:"document"`
	DocumentType models.CustomerDocumentType `json:"document_type"`
	Email        string                      `json:"email"`
	Phone        string                      `json:"phone"`
	Status       models.CustomerStatus       `json:"status"`
	URL          string                      `json:"url"`
	CreatedAt    time.Time                   `json:"created_at"`
	UpdatedAt    time.Time                   `json:"updated_at"`
}

func (request CustomerRequest) Validate() []errors.ApiErrorItem {
	errs := make([]errors.ApiErrorItem, 0)

	if utils.IsBlank(request.Name) || len(strings.TrimSpace(request.Name)) > 100 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_NAME_INVALID"))
	}

	if utils.IsBlank(request.Slug) || len(strings.TrimSpace(request.Slug)) > 50 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_SLUG_INVALID"))
	}

	if utils.IsBlank(request.Document) || len(strings.TrimSpace(request.Document)) > 100 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_DOCUMENT_INVALID"))
	}

	if !request.DocumentType.IsValid() {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_DOCUMENT_TYPE_INVALID"))
	}

	if request.DocumentType == models.CustomerDocumentTypeCPF && !utils.IsValidCPF(request.Document) {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_CPF_INVALID"))
	}

	if request.DocumentType == models.CustomerDocumentTypeCNPJ && !utils.IsValidCNPJ(request.Document) {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_CNPJ_INVALID"))
	}

	if utils.IsBlank(request.Email) || len(strings.TrimSpace(request.Email)) > 100 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_EMAIL_INVALID"))
	}

	if len(strings.TrimSpace(request.Phone)) > 50 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_PHONE_INVALID"))
	}

	if !request.Status.IsValid() {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_STATUS_INVALID"))
	}

	if len(strings.TrimSpace(request.URL)) > 500 {
		errs = append(errs, errors.InvalidFieldError("CUSTOMER_URL_INVALID"))
	}

	return errs
}

func CustomerResponseFromModel(customer models.Customer) CustomerResponse {
	return CustomerResponse{
		ID:           customer.ID,
		Name:         customer.Name,
		Slug:         customer.Slug,
		Document:     customer.Document,
		DocumentType: customer.DocumentType,
		Email:        customer.Email,
		Phone:        customer.Phone,
		Status:       customer.Status,
		URL:          customer.URL,
		CreatedAt:    customer.CreatedAt,
		UpdatedAt:    customer.UpdatedAt,
	}
}
