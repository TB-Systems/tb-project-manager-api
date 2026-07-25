package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CustomerProject interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.CustomerProjectResponse], errors.ApiError)
	FindByID(ctx context.Context, id string) (dto.CustomerProjectResponse, errors.ApiError)
	Create(ctx context.Context, request dto.CustomerProjectRequest) (dto.CustomerProjectResponse, errors.ApiError)
	Update(ctx context.Context, id string, request dto.CustomerProjectRequest) (dto.CustomerProjectResponse, errors.ApiError)
	Delete(ctx context.Context, id string) errors.ApiError
}

type customerProject struct {
	repository     repositories.CustomerProject
	statusSync     ProjectStatusSync
	billingSync    CustomerProjectBillingSync
	customerStatus CustomerStatusSync
}

func NewCustomerProjectService(repository repositories.CustomerProject, statusSync ProjectStatusSync, syncs ...interface{}) CustomerProject {
	service := customerProject{repository: repository, statusSync: statusSync}
	for _, sync := range syncs {
		switch typedSync := sync.(type) {
		case CustomerProjectBillingSync:
			service.billingSync = typedSync
		case CustomerStatusSync:
			service.customerStatus = typedSync
		}
	}

	return service
}

func (c customerProject) List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.CustomerProjectResponse], errors.ApiError) {
	if apiErr := c.syncBilling(ctx); apiErr != nil {
		return commonsmodels.PaginatedResponse[dto.CustomerProjectResponse]{}, apiErr
	}

	customerProjects, total, err := c.repository.List(ctx, params)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.CustomerProjectResponse]{}, internalCustomerProjectError("LIST_CUSTOMER_PROJECTS_FAILED")
	}

	items := make([]dto.CustomerProjectResponse, 0, len(customerProjects))
	for _, customerProject := range customerProjects {
		items = append(items, dto.CustomerProjectResponseFromModel(customerProject))
	}

	return commonsmodels.PaginatedResponse[dto.CustomerProjectResponse]{
		Items:     items,
		PageCount: pageCount(total, params.Limit),
		Page:      int64(params.Page),
	}, nil
}

func (c customerProject) FindByID(ctx context.Context, id string) (dto.CustomerProjectResponse, errors.ApiError) {
	customerProjectID, apiErr := parseCustomerProjectID(id)
	if apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	if apiErr := c.syncBilling(ctx); apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	customerProject, err := c.repository.FindByID(ctx, customerProjectID)
	if err != nil {
		return dto.CustomerProjectResponse{}, customerProjectRepositoryError(err, "FIND_CUSTOMER_PROJECT_FAILED")
	}

	return dto.CustomerProjectResponseFromModel(customerProject), nil
}

func (c customerProject) Create(ctx context.Context, request dto.CustomerProjectRequest) (dto.CustomerProjectResponse, errors.ApiError) {
	customerProject := customerProjectFromRequest(request)

	exists, err := c.repository.LinkExists(ctx, customerProject.ProjectID, customerProject.CustomerID, nil)
	if err != nil {
		return dto.CustomerProjectResponse{}, internalCustomerProjectError("CHECK_CUSTOMER_PROJECT_LINK_FAILED")
	}
	if exists {
		return dto.CustomerProjectResponse{}, customerProjectConflictError("CUSTOMER_PROJECT_ALREADY_EXISTS")
	}

	createdCustomerProject, err := c.repository.Create(ctx, customerProject)
	if err != nil {
		return dto.CustomerProjectResponse{}, internalCustomerProjectError("CREATE_CUSTOMER_PROJECT_FAILED")
	}

	if apiErr := c.statusSync.Sync(ctx, customerProject.ProjectID); apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	if apiErr := c.syncCustomerStatus(ctx, customerProject.CustomerID); apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	return dto.CustomerProjectResponseFromModel(createdCustomerProject), nil
}

func (c customerProject) Update(ctx context.Context, id string, request dto.CustomerProjectRequest) (dto.CustomerProjectResponse, errors.ApiError) {
	customerProjectID, apiErr := parseCustomerProjectID(id)
	if apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	currentCustomerProject, err := c.repository.FindByID(ctx, customerProjectID)
	if err != nil {
		return dto.CustomerProjectResponse{}, customerProjectRepositoryError(err, "FIND_CUSTOMER_PROJECT_FAILED")
	}

	customerProject := customerProjectFromRequest(request)
	customerProject.ID = customerProjectID
	customerProject.Status = currentCustomerProject.Status
	customerProject.StartedAt = currentCustomerProject.StartedAt
	customerProject.ClosedAt = currentCustomerProject.ClosedAt

	exists, err := c.repository.LinkExists(ctx, customerProject.ProjectID, customerProject.CustomerID, &customerProjectID)
	if err != nil {
		return dto.CustomerProjectResponse{}, internalCustomerProjectError("CHECK_CUSTOMER_PROJECT_LINK_FAILED")
	}
	if exists {
		return dto.CustomerProjectResponse{}, customerProjectConflictError("CUSTOMER_PROJECT_ALREADY_EXISTS")
	}

	updatedCustomerProject, err := c.repository.Update(ctx, customerProject)
	if err != nil {
		return dto.CustomerProjectResponse{}, internalCustomerProjectError("UPDATE_CUSTOMER_PROJECT_FAILED")
	}

	if apiErr := c.statusSync.Sync(ctx, currentCustomerProject.ProjectID); apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	if currentCustomerProject.ProjectID != customerProject.ProjectID {
		if apiErr := c.statusSync.Sync(ctx, customerProject.ProjectID); apiErr != nil {
			return dto.CustomerProjectResponse{}, apiErr
		}
	}

	if apiErr := c.syncCustomerStatus(ctx, currentCustomerProject.CustomerID); apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	if currentCustomerProject.CustomerID != customerProject.CustomerID {
		if apiErr := c.syncCustomerStatus(ctx, customerProject.CustomerID); apiErr != nil {
			return dto.CustomerProjectResponse{}, apiErr
		}
	}

	return dto.CustomerProjectResponseFromModel(updatedCustomerProject), nil
}

func (c customerProject) Delete(ctx context.Context, id string) errors.ApiError {
	customerProjectID, apiErr := parseCustomerProjectID(id)
	if apiErr != nil {
		return apiErr
	}

	customerProject, err := c.repository.FindByID(ctx, customerProjectID)
	if err != nil {
		return customerProjectRepositoryError(err, "FIND_CUSTOMER_PROJECT_FAILED")
	}

	if err := c.repository.Delete(ctx, customerProjectID); err != nil {
		return internalCustomerProjectError("DELETE_CUSTOMER_PROJECT_FAILED")
	}

	if apiErr := c.statusSync.Sync(ctx, customerProject.ProjectID); apiErr != nil {
		return apiErr
	}

	if apiErr := c.syncCustomerStatus(ctx, customerProject.CustomerID); apiErr != nil {
		return apiErr
	}

	return nil
}

func (c customerProject) syncBilling(ctx context.Context) errors.ApiError {
	if c.billingSync == nil {
		return nil
	}

	return c.billingSync.SyncOverdue(ctx)
}

func (c customerProject) syncCustomerStatus(ctx context.Context, customerID uuid.UUID) errors.ApiError {
	if c.customerStatus == nil {
		return nil
	}

	return c.customerStatus.Sync(ctx, customerID)
}

func customerProjectFromRequest(request dto.CustomerProjectRequest) models.CustomerProject {
	now := time.Now().UTC()
	setupFirstHalf, setupSecondHalf := splitSetupValue(request.ProjectValue)

	return models.CustomerProject{
		ProjectID:  request.ProjectID,
		CustomerID: request.CustomerID,
		Status:     models.CustomerProjectStatusActive,
		StartedAt:  now,
		Terms: []models.CustomerProjectTerm{
			{
				SetupValue:   request.ProjectValue,
				MonthlyValue: request.MonthlyValue,
				DueDay:       request.DueDay,
				StartsAt:     now,
				Active:       true,
			},
		},
		Invoices: setupInvoices(setupFirstHalf, setupSecondHalf, now),
	}
}

func splitSetupValue(value int) (int, int) {
	firstHalf := (value + 1) / 2
	return firstHalf, value - firstHalf
}

func setupInvoices(firstHalf int, secondHalf int, now time.Time) []models.CustomerProjectInvoice {
	invoices := make([]models.CustomerProjectInvoice, 0, 2)
	if firstHalf > 0 {
		invoices = append(invoices, models.CustomerProjectInvoice{
			Type:    models.CustomerProjectInvoiceTypeSetupFirstHalf,
			Amount:  firstHalf,
			DueDate: beginningOfDay(now),
			Status:  models.CustomerProjectInvoiceStatusOpen,
		})
	}

	if secondHalf > 0 {
		invoices = append(invoices, models.CustomerProjectInvoice{
			Type:    models.CustomerProjectInvoiceTypeSetupSecondHalf,
			Amount:  secondHalf,
			DueDate: beginningOfDay(now),
			Status:  models.CustomerProjectInvoiceStatusOpen,
		})
	}

	return invoices
}

func beginningOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}

func parseCustomerProjectID(id string) (uuid.UUID, errors.ApiError) {
	customerProjectID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("CUSTOMER_PROJECT_ID_INVALID"),
		)
	}

	return customerProjectID, nil
}

func customerProjectRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("CUSTOMER_PROJECT_NOT_FOUND"),
		)
	}

	return internalCustomerProjectError(detail)
}

func customerProjectConflictError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusConflict,
		errors.BadRequestError(detail),
	)
}

func internalCustomerProjectError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}
