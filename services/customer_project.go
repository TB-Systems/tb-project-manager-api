package services

import (
	"context"
	stderrors "errors"
	"net/http"

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
	repository repositories.CustomerProject
}

func NewCustomerProjectService(repository repositories.CustomerProject) CustomerProject {
	return customerProject{repository: repository}
}

func (c customerProject) List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.CustomerProjectResponse], errors.ApiError) {
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

	return dto.CustomerProjectResponseFromModel(createdCustomerProject), nil
}

func (c customerProject) Update(ctx context.Context, id string, request dto.CustomerProjectRequest) (dto.CustomerProjectResponse, errors.ApiError) {
	customerProjectID, apiErr := parseCustomerProjectID(id)
	if apiErr != nil {
		return dto.CustomerProjectResponse{}, apiErr
	}

	if _, err := c.repository.FindByID(ctx, customerProjectID); err != nil {
		return dto.CustomerProjectResponse{}, customerProjectRepositoryError(err, "FIND_CUSTOMER_PROJECT_FAILED")
	}

	customerProject := customerProjectFromRequest(request)
	customerProject.ID = customerProjectID

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

	return dto.CustomerProjectResponseFromModel(updatedCustomerProject), nil
}

func (c customerProject) Delete(ctx context.Context, id string) errors.ApiError {
	customerProjectID, apiErr := parseCustomerProjectID(id)
	if apiErr != nil {
		return apiErr
	}

	if _, err := c.repository.FindByID(ctx, customerProjectID); err != nil {
		return customerProjectRepositoryError(err, "FIND_CUSTOMER_PROJECT_FAILED")
	}

	if err := c.repository.Delete(ctx, customerProjectID); err != nil {
		return internalCustomerProjectError("DELETE_CUSTOMER_PROJECT_FAILED")
	}

	return nil
}

func customerProjectFromRequest(request dto.CustomerProjectRequest) models.CustomerProject {
	return models.CustomerProject{
		ProjectID:            request.ProjectID,
		CustomerID:           request.CustomerID,
		ProjectValue:         request.ProjectValue,
		MonthlyValue:         request.MonthlyValue,
		DueDay:               request.DueDay,
		ProjectPaymentStatus: request.ProjectPaymentStatus,
		LastPayment:          request.LastPayment,
	}
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
