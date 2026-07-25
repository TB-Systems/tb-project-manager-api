package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerProjectInvoice interface {
	Create(ctx context.Context, customerProjectID string, request dto.CustomerProjectInvoiceRequest) (dto.CustomerProjectInvoiceResponse, errors.ApiError)
	Update(ctx context.Context, id string, request dto.CustomerProjectInvoiceRequest) (dto.CustomerProjectInvoiceResponse, errors.ApiError)
	Pay(ctx context.Context, id string, request dto.CustomerProjectInvoicePayRequest) (dto.CustomerProjectInvoiceResponse, errors.ApiError)
	Unpay(ctx context.Context, id string) (dto.CustomerProjectInvoiceResponse, errors.ApiError)
}

type CustomerProjectBillingSync interface {
	SyncOverdue(ctx context.Context) errors.ApiError
}

type customerProjectInvoice struct {
	repository     repositories.CustomerProjectInvoice
	customerStatus CustomerStatusSync
	now            func() time.Time
}

func NewCustomerProjectInvoiceService(repository repositories.CustomerProjectInvoice, customerStatus ...CustomerStatusSync) CustomerProjectInvoice {
	service := customerProjectInvoice{repository: repository, now: time.Now}
	if len(customerStatus) > 0 {
		service.customerStatus = customerStatus[0]
	}

	return service
}

func NewCustomerProjectBillingSyncService(repository repositories.CustomerProjectInvoice, customerStatus ...CustomerStatusSync) CustomerProjectBillingSync {
	service := customerProjectInvoice{repository: repository, now: time.Now}
	if len(customerStatus) > 0 {
		service.customerStatus = customerStatus[0]
	}

	return service
}

func (c customerProjectInvoice) Create(ctx context.Context, customerProjectID string, request dto.CustomerProjectInvoiceRequest) (dto.CustomerProjectInvoiceResponse, errors.ApiError) {
	parsedCustomerProjectID, apiErr := parseCustomerProjectID(customerProjectID)
	if apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	invoice := invoiceFromRequest(request)
	invoice.CustomerProjectID = parsedCustomerProjectID

	createdInvoice, err := c.repository.Create(ctx, invoice)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, internalCustomerProjectInvoiceError("CREATE_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	if apiErr := c.SyncOverdue(ctx); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	if apiErr := c.syncCustomerStatus(ctx, []uuid.UUID{createdInvoice.CustomerProject.CustomerID}); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	return dto.CustomerProjectInvoiceResponseFromModel(createdInvoice), nil
}

func (c customerProjectInvoice) Update(ctx context.Context, id string, request dto.CustomerProjectInvoiceRequest) (dto.CustomerProjectInvoiceResponse, errors.ApiError) {
	invoiceID, apiErr := parseCustomerProjectInvoiceID(id)
	if apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	currentInvoice, err := c.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, customerProjectInvoiceRepositoryError(err, "FIND_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	invoice := invoiceFromRequest(request)
	invoice.ID = invoiceID
	invoice.CustomerProjectID = currentInvoice.CustomerProjectID

	updatedInvoice, err := c.repository.Update(ctx, invoice)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, internalCustomerProjectInvoiceError("UPDATE_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	if apiErr := c.SyncOverdue(ctx); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	if apiErr := c.syncCustomerStatus(ctx, []uuid.UUID{updatedInvoice.CustomerProject.CustomerID}); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	return dto.CustomerProjectInvoiceResponseFromModel(updatedInvoice), nil
}

func invoiceFromRequest(request dto.CustomerProjectInvoiceRequest) models.CustomerProjectInvoice {
	return models.CustomerProjectInvoice{
		Type:           request.Type,
		ReferenceMonth: normalizeReferenceMonth(request.ReferenceMonth),
		Amount:         request.Amount,
		DueDate:        beginningOfDay(request.DueDate.UTC()),
		Status:         request.Status,
		PaidAt:         normalizePaidAt(request.Status, request.PaidAt),
	}
}

func (c customerProjectInvoice) Pay(ctx context.Context, id string, request dto.CustomerProjectInvoicePayRequest) (dto.CustomerProjectInvoiceResponse, errors.ApiError) {
	invoiceID, apiErr := parseCustomerProjectInvoiceID(id)
	if apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	if apiErr := c.SyncOverdue(ctx); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	invoice, err := c.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, customerProjectInvoiceRepositoryError(err, "FIND_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	if invoice.Status == models.CustomerProjectInvoiceStatusCancelled {
		return dto.CustomerProjectInvoiceResponse{}, customerProjectInvoiceBadRequestError("CUSTOMER_PROJECT_INVOICE_CANCELLED")
	}

	if invoice.Status == models.CustomerProjectInvoiceStatusPaid {
		return dto.CustomerProjectInvoiceResponseFromModel(invoice), nil
	}

	paidAt := c.now().UTC()
	if request.PaidAt != nil {
		paidAt = request.PaidAt.UTC()
	}

	paidInvoice, err := c.repository.Pay(ctx, invoiceID, paidAt)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, internalCustomerProjectInvoiceError("PAY_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	if apiErr := c.syncCustomerStatus(ctx, []uuid.UUID{paidInvoice.CustomerProject.CustomerID}); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	return dto.CustomerProjectInvoiceResponseFromModel(paidInvoice), nil
}

func (c customerProjectInvoice) Unpay(ctx context.Context, id string) (dto.CustomerProjectInvoiceResponse, errors.ApiError) {
	invoiceID, apiErr := parseCustomerProjectInvoiceID(id)
	if apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	invoice, err := c.repository.FindByID(ctx, invoiceID)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, customerProjectInvoiceRepositoryError(err, "FIND_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	status := models.CustomerProjectInvoiceStatusOpen
	if invoice.DueDate.Before(beginningOfDay(c.now().UTC())) {
		status = models.CustomerProjectInvoiceStatusOverdue
	}

	unpaidInvoice, err := c.repository.Unpay(ctx, invoiceID, status)
	if err != nil {
		return dto.CustomerProjectInvoiceResponse{}, internalCustomerProjectInvoiceError("UNPAY_CUSTOMER_PROJECT_INVOICE_FAILED")
	}

	if apiErr := c.syncCustomerStatus(ctx, []uuid.UUID{unpaidInvoice.CustomerProject.CustomerID}); apiErr != nil {
		return dto.CustomerProjectInvoiceResponse{}, apiErr
	}

	return dto.CustomerProjectInvoiceResponseFromModel(unpaidInvoice), nil
}

func (c customerProjectInvoice) SyncOverdue(ctx context.Context) errors.ApiError {
	customerIDs, err := c.repository.MarkOverdue(ctx, beginningOfDay(c.now().UTC()))
	if err != nil {
		return internalCustomerProjectInvoiceError("SYNC_CUSTOMER_PROJECT_INVOICES_OVERDUE_FAILED")
	}

	if apiErr := c.syncCustomerStatus(ctx, customerIDs); apiErr != nil {
		return apiErr
	}

	return nil
}

func normalizeReferenceMonth(referenceMonth *time.Time) *time.Time {
	if referenceMonth == nil {
		return nil
	}

	normalized := time.Date(referenceMonth.Year(), referenceMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	return &normalized
}

func normalizePaidAt(status models.CustomerProjectInvoiceStatus, paidAt *time.Time) *time.Time {
	if status != models.CustomerProjectInvoiceStatusPaid || paidAt == nil {
		return nil
	}

	normalized := paidAt.UTC()
	return &normalized
}

func (c customerProjectInvoice) syncCustomerStatus(ctx context.Context, customerIDs []uuid.UUID) errors.ApiError {
	if c.customerStatus == nil {
		return nil
	}

	return c.customerStatus.SyncMany(ctx, customerIDs)
}

func parseCustomerProjectInvoiceID(id string) (uuid.UUID, errors.ApiError) {
	invoiceID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("CUSTOMER_PROJECT_INVOICE_ID_INVALID"),
		)
	}

	return invoiceID, nil
}

func customerProjectInvoiceRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("CUSTOMER_PROJECT_INVOICE_NOT_FOUND"),
		)
	}

	return internalCustomerProjectInvoiceError(detail)
}

func customerProjectInvoiceBadRequestError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusBadRequest,
		errors.BadRequestError(detail),
	)
}

func internalCustomerProjectInvoiceError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}
