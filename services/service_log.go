package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/TB-Systems/go-commons/commonsmodels"
	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ServiceLog interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ServiceLogResponse], errors.ApiError)
	FindByID(ctx context.Context, id string) (dto.ServiceLogResponse, errors.ApiError)
	Create(ctx context.Context, request dto.ServiceLogRequest) (dto.ServiceLogResponse, errors.ApiError)
}

type serviceLog struct {
	repository repositories.ServiceLog
}

func NewServiceLogService(repository repositories.ServiceLog) ServiceLog {
	return serviceLog{repository: repository}
}

func (s serviceLog) List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ServiceLogResponse], errors.ApiError) {
	serviceLogs, total, err := s.repository.List(ctx, params)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.ServiceLogResponse]{}, internalServiceLogError("LIST_SERVICE_LOGS_FAILED")
	}

	items := make([]dto.ServiceLogResponse, 0, len(serviceLogs))
	for _, serviceLog := range serviceLogs {
		items = append(items, dto.ServiceLogResponseFromModel(serviceLog))
	}

	return commonsmodels.PaginatedResponse[dto.ServiceLogResponse]{
		Items:     items,
		PageCount: pageCount(total, params.Limit),
		Page:      int64(params.Page),
	}, nil
}

func (s serviceLog) FindByID(ctx context.Context, id string) (dto.ServiceLogResponse, errors.ApiError) {
	serviceLogID, apiErr := parseServiceLogID(id)
	if apiErr != nil {
		return dto.ServiceLogResponse{}, apiErr
	}

	serviceLog, err := s.repository.FindByID(ctx, serviceLogID)
	if err != nil {
		return dto.ServiceLogResponse{}, serviceLogRepositoryError(err, "FIND_SERVICE_LOG_FAILED")
	}

	return dto.ServiceLogResponseFromModel(serviceLog), nil
}

func (s serviceLog) Create(ctx context.Context, request dto.ServiceLogRequest) (dto.ServiceLogResponse, errors.ApiError) {
	serviceLog := serviceLogFromRequest(request)

	createdServiceLog, err := s.repository.Create(ctx, serviceLog)
	if err != nil {
		return dto.ServiceLogResponse{}, internalServiceLogError("CREATE_SERVICE_LOG_FAILED")
	}

	return dto.ServiceLogResponseFromModel(createdServiceLog), nil
}

func serviceLogFromRequest(request dto.ServiceLogRequest) models.ServiceLog {
	return models.ServiceLog{
		ProjectServiceID: request.ProjectServiceID,
		Level:            request.Level,
		Event:            strings.TrimSpace(request.Event),
		Message:          request.Message,
		Time:             request.Time,
	}
}

func parseServiceLogID(id string) (uuid.UUID, errors.ApiError) {
	serviceLogID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("SERVICE_LOG_ID_INVALID"),
		)
	}

	return serviceLogID, nil
}

func serviceLogRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("SERVICE_LOG_NOT_FOUND"),
		)
	}

	return internalServiceLogError(detail)
}

func internalServiceLogError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}
