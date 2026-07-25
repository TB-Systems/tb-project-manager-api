package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Customer interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.CustomerResponse], errors.ApiError)
	FindByID(ctx context.Context, id string) (dto.CustomerResponse, errors.ApiError)
	Create(ctx context.Context, request dto.CustomerRequest) (dto.CustomerResponse, errors.ApiError)
	Update(ctx context.Context, id string, request dto.CustomerRequest) (dto.CustomerResponse, errors.ApiError)
	Delete(ctx context.Context, id string) errors.ApiError
}

type customer struct {
	repository repositories.Customer
	statusSync CustomerStatusSync
}

func NewCustomerService(repository repositories.Customer, statusSync ...CustomerStatusSync) Customer {
	service := customer{repository: repository}
	if len(statusSync) > 0 {
		service.statusSync = statusSync[0]
	}

	return service
}

func (c customer) List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.CustomerResponse], errors.ApiError) {
	customers, total, err := c.repository.List(ctx, params)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.CustomerResponse]{}, internalCustomerError("LIST_CUSTOMERS_FAILED")
	}

	for _, customer := range customers {
		if apiErr := c.syncStatus(ctx, customer.ID); apiErr != nil {
			return commonsmodels.PaginatedResponse[dto.CustomerResponse]{}, apiErr
		}
	}

	customers, _, err = c.repository.List(ctx, params)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.CustomerResponse]{}, internalCustomerError("LIST_CUSTOMERS_FAILED")
	}

	items := make([]dto.CustomerResponse, 0, len(customers))
	for _, customer := range customers {
		items = append(items, dto.CustomerResponseFromModel(customer))
	}

	return commonsmodels.PaginatedResponse[dto.CustomerResponse]{
		Items:     items,
		PageCount: pageCount(total, params.Limit),
		Page:      int64(params.Page),
	}, nil
}

func (c customer) FindByID(ctx context.Context, id string) (dto.CustomerResponse, errors.ApiError) {
	customerID, apiErr := parseCustomerID(id)
	if apiErr != nil {
		return dto.CustomerResponse{}, apiErr
	}

	if apiErr := c.syncStatus(ctx, customerID); apiErr != nil {
		return dto.CustomerResponse{}, apiErr
	}

	customer, err := c.repository.FindByID(ctx, customerID)
	if err != nil {
		return dto.CustomerResponse{}, customerRepositoryError(err, "FIND_CUSTOMER_FAILED")
	}

	return dto.CustomerResponseFromModel(customer), nil
}

func (c customer) Create(ctx context.Context, request dto.CustomerRequest) (dto.CustomerResponse, errors.ApiError) {
	customer := customerFromRequest(request)

	if apiErr := c.validateUniqueFields(ctx, customer, nil); apiErr != nil {
		return dto.CustomerResponse{}, apiErr
	}

	createdCustomer, err := c.repository.Create(ctx, customer)
	if err != nil {
		return dto.CustomerResponse{}, internalCustomerError("CREATE_CUSTOMER_FAILED")
	}

	if c.statusSync != nil {
		if apiErr := c.syncStatus(ctx, createdCustomer.ID); apiErr != nil {
			return dto.CustomerResponse{}, apiErr
		}

		syncedCustomer, err := c.repository.FindByID(ctx, createdCustomer.ID)
		if err != nil {
			return dto.CustomerResponse{}, customerRepositoryError(err, "FIND_CUSTOMER_FAILED")
		}

		return dto.CustomerResponseFromModel(syncedCustomer), nil
	}

	return dto.CustomerResponseFromModel(createdCustomer), nil
}

func (c customer) Update(ctx context.Context, id string, request dto.CustomerRequest) (dto.CustomerResponse, errors.ApiError) {
	customerID, apiErr := parseCustomerID(id)
	if apiErr != nil {
		return dto.CustomerResponse{}, apiErr
	}

	if _, err := c.repository.FindByID(ctx, customerID); err != nil {
		return dto.CustomerResponse{}, customerRepositoryError(err, "FIND_CUSTOMER_FAILED")
	}

	customer := customerFromRequest(request)
	customer.ID = customerID

	if apiErr := c.validateUniqueFields(ctx, customer, &customerID); apiErr != nil {
		return dto.CustomerResponse{}, apiErr
	}

	updatedCustomer, err := c.repository.Update(ctx, customer)
	if err != nil {
		return dto.CustomerResponse{}, internalCustomerError("UPDATE_CUSTOMER_FAILED")
	}

	if c.statusSync != nil {
		if apiErr := c.syncStatus(ctx, updatedCustomer.ID); apiErr != nil {
			return dto.CustomerResponse{}, apiErr
		}

		syncedCustomer, err := c.repository.FindByID(ctx, updatedCustomer.ID)
		if err != nil {
			return dto.CustomerResponse{}, customerRepositoryError(err, "FIND_CUSTOMER_FAILED")
		}

		return dto.CustomerResponseFromModel(syncedCustomer), nil
	}

	return dto.CustomerResponseFromModel(updatedCustomer), nil
}

func (c customer) Delete(ctx context.Context, id string) errors.ApiError {
	customerID, apiErr := parseCustomerID(id)
	if apiErr != nil {
		return apiErr
	}

	if _, err := c.repository.FindByID(ctx, customerID); err != nil {
		return customerRepositoryError(err, "FIND_CUSTOMER_FAILED")
	}

	if err := c.repository.Delete(ctx, customerID); err != nil {
		return internalCustomerError("DELETE_CUSTOMER_FAILED")
	}

	return nil
}

func (c customer) validateUniqueFields(ctx context.Context, customer models.Customer, exceptID *uuid.UUID) errors.ApiError {
	exists, err := c.repository.SlugExists(ctx, customer.Slug, exceptID)
	if err != nil {
		return internalCustomerError("CHECK_CUSTOMER_SLUG_FAILED")
	}
	if exists {
		return customerConflictError("CUSTOMER_SLUG_ALREADY_EXISTS")
	}

	exists, err = c.repository.DocumentExists(ctx, customer.Document, exceptID)
	if err != nil {
		return internalCustomerError("CHECK_CUSTOMER_DOCUMENT_FAILED")
	}
	if exists {
		return customerConflictError("CUSTOMER_DOCUMENT_ALREADY_EXISTS")
	}

	exists, err = c.repository.EmailExists(ctx, customer.Email, exceptID)
	if err != nil {
		return internalCustomerError("CHECK_CUSTOMER_EMAIL_FAILED")
	}
	if exists {
		return customerConflictError("CUSTOMER_EMAIL_ALREADY_EXISTS")
	}

	return nil
}

func (c customer) syncStatus(ctx context.Context, customerID uuid.UUID) errors.ApiError {
	if c.statusSync == nil {
		return nil
	}

	return c.statusSync.Sync(ctx, customerID)
}

func customerFromRequest(request dto.CustomerRequest) models.Customer {
	return models.Customer{
		Name:         strings.TrimSpace(request.Name),
		Slug:         strings.TrimSpace(request.Slug),
		Document:     utils.NormalizeDocument(request.Document),
		DocumentType: request.DocumentType,
		Email:        strings.TrimSpace(request.Email),
		Phone:        strings.TrimSpace(request.Phone),
		Status:       request.Status,
		URL:          strings.TrimSpace(request.URL),
	}
}

func parseCustomerID(id string) (uuid.UUID, errors.ApiError) {
	customerID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("CUSTOMER_ID_INVALID"),
		)
	}

	return customerID, nil
}

func customerRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("CUSTOMER_NOT_FOUND"),
		)
	}

	return internalCustomerError(detail)
}

func customerConflictError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusConflict,
		errors.BadRequestError(detail),
	)
}

func internalCustomerError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}
