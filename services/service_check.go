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

type ServiceCheck interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ServiceCheckResponse], errors.ApiError)
	FindByID(ctx context.Context, id string) (dto.ServiceCheckResponse, errors.ApiError)
	Create(ctx context.Context, request dto.ServiceCheckRequest) (dto.ServiceCheckResponse, errors.ApiError)
	Update(ctx context.Context, id string, request dto.ServiceCheckRequest) (dto.ServiceCheckResponse, errors.ApiError)
	Delete(ctx context.Context, id string) errors.ApiError
}

type serviceCheck struct {
	repository repositories.ServiceCheck
}

func NewServiceCheckService(repository repositories.ServiceCheck) ServiceCheck {
	return serviceCheck{repository: repository}
}

func (s serviceCheck) List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ServiceCheckResponse], errors.ApiError) {
	serviceChecks, total, err := s.repository.List(ctx, params)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.ServiceCheckResponse]{}, internalServiceCheckError("LIST_SERVICE_CHECKS_FAILED")
	}

	items := make([]dto.ServiceCheckResponse, 0, len(serviceChecks))
	for _, serviceCheck := range serviceChecks {
		items = append(items, dto.ServiceCheckResponseFromModel(serviceCheck))
	}

	return commonsmodels.PaginatedResponse[dto.ServiceCheckResponse]{
		Items:     items,
		PageCount: pageCount(total, params.Limit),
		Page:      int64(params.Page),
	}, nil
}

func (s serviceCheck) FindByID(ctx context.Context, id string) (dto.ServiceCheckResponse, errors.ApiError) {
	serviceCheckID, apiErr := parseServiceCheckID(id)
	if apiErr != nil {
		return dto.ServiceCheckResponse{}, apiErr
	}

	serviceCheck, err := s.repository.FindByID(ctx, serviceCheckID)
	if err != nil {
		return dto.ServiceCheckResponse{}, serviceCheckRepositoryError(err, "FIND_SERVICE_CHECK_FAILED")
	}

	return dto.ServiceCheckResponseFromModel(serviceCheck), nil
}

func (s serviceCheck) Create(ctx context.Context, request dto.ServiceCheckRequest) (dto.ServiceCheckResponse, errors.ApiError) {
	serviceCheck := serviceCheckFromRequest(request)

	createdServiceCheck, err := s.repository.Create(ctx, serviceCheck)
	if err != nil {
		return dto.ServiceCheckResponse{}, internalServiceCheckError("CREATE_SERVICE_CHECK_FAILED")
	}

	return dto.ServiceCheckResponseFromModel(createdServiceCheck), nil
}

func (s serviceCheck) Update(ctx context.Context, id string, request dto.ServiceCheckRequest) (dto.ServiceCheckResponse, errors.ApiError) {
	serviceCheckID, apiErr := parseServiceCheckID(id)
	if apiErr != nil {
		return dto.ServiceCheckResponse{}, apiErr
	}

	if _, err := s.repository.FindByID(ctx, serviceCheckID); err != nil {
		return dto.ServiceCheckResponse{}, serviceCheckRepositoryError(err, "FIND_SERVICE_CHECK_FAILED")
	}

	serviceCheck := serviceCheckFromRequest(request)
	serviceCheck.ID = serviceCheckID

	updatedServiceCheck, err := s.repository.Update(ctx, serviceCheck)
	if err != nil {
		return dto.ServiceCheckResponse{}, internalServiceCheckError("UPDATE_SERVICE_CHECK_FAILED")
	}

	return dto.ServiceCheckResponseFromModel(updatedServiceCheck), nil
}

func (s serviceCheck) Delete(ctx context.Context, id string) errors.ApiError {
	serviceCheckID, apiErr := parseServiceCheckID(id)
	if apiErr != nil {
		return apiErr
	}

	if _, err := s.repository.FindByID(ctx, serviceCheckID); err != nil {
		return serviceCheckRepositoryError(err, "FIND_SERVICE_CHECK_FAILED")
	}

	if err := s.repository.Delete(ctx, serviceCheckID); err != nil {
		return internalServiceCheckError("DELETE_SERVICE_CHECK_FAILED")
	}

	return nil
}

func serviceCheckFromRequest(request dto.ServiceCheckRequest) models.ServiceCheck {
	return models.ServiceCheck{
		ProjectServiceID: request.ProjectServiceID,
		Status:           request.Status,
		StatusCode:       request.StatusCode,
		ResponseTimeMS:   request.ResponseTimeMS,
		Message:          request.Message,
	}
}

func parseServiceCheckID(id string) (uuid.UUID, errors.ApiError) {
	serviceCheckID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("SERVICE_CHECK_ID_INVALID"),
		)
	}

	return serviceCheckID, nil
}

func serviceCheckRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("SERVICE_CHECK_NOT_FOUND"),
		)
	}

	return internalServiceCheckError(detail)
}

func internalServiceCheckError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}
