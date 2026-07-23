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

type ProjectService interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams, filter dto.ProjectServiceListFilter) (commonsmodels.PaginatedResponse[dto.ProjectServiceResponse], errors.ApiError)
	FindByID(ctx context.Context, id string) (dto.ProjectServiceResponse, errors.ApiError)
	Create(ctx context.Context, request dto.ProjectServiceCreateRequest) (dto.ProjectServiceResponse, errors.ApiError)
	Update(ctx context.Context, id string, request dto.ProjectServiceRequest) (dto.ProjectServiceResponse, errors.ApiError)
	Delete(ctx context.Context, id string) errors.ApiError
}

type projectService struct {
	repository repositories.ProjectService
	statusSync ProjectStatusSync
}

func NewProjectServiceService(repository repositories.ProjectService, statusSync ProjectStatusSync) ProjectService {
	return projectService{repository: repository, statusSync: statusSync}
}

func (p projectService) List(ctx context.Context, params commonsmodels.PaginatedParams, filter dto.ProjectServiceListFilter) (commonsmodels.PaginatedResponse[dto.ProjectServiceResponse], errors.ApiError) {
	repositoryFilter, apiErr := projectServiceRepositoryFilterFromListFilter(filter)
	if apiErr != nil {
		return commonsmodels.PaginatedResponse[dto.ProjectServiceResponse]{}, apiErr
	}

	projectServices, total, err := p.repository.List(ctx, params, repositoryFilter)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.ProjectServiceResponse]{}, internalProjectServiceError("LIST_PROJECT_SERVICES_FAILED")
	}

	items := make([]dto.ProjectServiceResponse, 0, len(projectServices))
	for _, projectService := range projectServices {
		items = append(items, dto.ProjectServiceResponseFromModel(projectService))
	}

	return commonsmodels.PaginatedResponse[dto.ProjectServiceResponse]{
		Items:     items,
		PageCount: pageCount(total, params.Limit),
		Page:      int64(params.Page),
	}, nil
}

func projectServiceRepositoryFilterFromListFilter(filter dto.ProjectServiceListFilter) (repositories.ProjectServiceFilter, errors.ApiError) {
	var repositoryFilter repositories.ProjectServiceFilter
	projectID := strings.TrimSpace(filter.ProjectID)
	if projectID == "" {
		return repositoryFilter, nil
	}

	parsedProjectID, err := uuid.Parse(projectID)
	if err != nil {
		return repositories.ProjectServiceFilter{}, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("PROJECT_ID_INVALID"),
		)
	}

	repositoryFilter.ProjectID = &parsedProjectID
	return repositoryFilter, nil
}

func (p projectService) FindByID(ctx context.Context, id string) (dto.ProjectServiceResponse, errors.ApiError) {
	projectServiceID, apiErr := parseProjectServiceID(id)
	if apiErr != nil {
		return dto.ProjectServiceResponse{}, apiErr
	}

	projectService, err := p.repository.FindByID(ctx, projectServiceID)
	if err != nil {
		return dto.ProjectServiceResponse{}, projectServiceRepositoryError(err, "FIND_PROJECT_SERVICE_FAILED")
	}

	return dto.ProjectServiceResponseFromModel(projectService), nil
}

func (p projectService) Create(ctx context.Context, request dto.ProjectServiceCreateRequest) (dto.ProjectServiceResponse, errors.ApiError) {
	projectService := projectServiceFromCreateRequest(request)
	projectService.Status = models.ProjectStatusBacklog

	createdProjectService, err := p.repository.Create(ctx, projectService)
	if err != nil {
		return dto.ProjectServiceResponse{}, internalProjectServiceError("CREATE_PROJECT_SERVICE_FAILED")
	}

	if apiErr := p.statusSync.Sync(ctx, projectService.ProjectID); apiErr != nil {
		return dto.ProjectServiceResponse{}, apiErr
	}

	return dto.ProjectServiceResponseFromModel(createdProjectService), nil
}

func (p projectService) Update(ctx context.Context, id string, request dto.ProjectServiceRequest) (dto.ProjectServiceResponse, errors.ApiError) {
	projectServiceID, apiErr := parseProjectServiceID(id)
	if apiErr != nil {
		return dto.ProjectServiceResponse{}, apiErr
	}

	currentProjectService, err := p.repository.FindByID(ctx, projectServiceID)
	if err != nil {
		return dto.ProjectServiceResponse{}, projectServiceRepositoryError(err, "FIND_PROJECT_SERVICE_FAILED")
	}

	projectService := projectServiceFromRequest(request)
	projectService.ID = projectServiceID

	updatedProjectService, err := p.repository.Update(ctx, projectService)
	if err != nil {
		return dto.ProjectServiceResponse{}, internalProjectServiceError("UPDATE_PROJECT_SERVICE_FAILED")
	}

	if apiErr := p.statusSync.Sync(ctx, currentProjectService.ProjectID); apiErr != nil {
		return dto.ProjectServiceResponse{}, apiErr
	}

	if currentProjectService.ProjectID != projectService.ProjectID {
		if apiErr := p.statusSync.Sync(ctx, projectService.ProjectID); apiErr != nil {
			return dto.ProjectServiceResponse{}, apiErr
		}
	}

	return dto.ProjectServiceResponseFromModel(updatedProjectService), nil
}

func (p projectService) Delete(ctx context.Context, id string) errors.ApiError {
	projectServiceID, apiErr := parseProjectServiceID(id)
	if apiErr != nil {
		return apiErr
	}

	projectService, err := p.repository.FindByID(ctx, projectServiceID)
	if err != nil {
		return projectServiceRepositoryError(err, "FIND_PROJECT_SERVICE_FAILED")
	}

	if err := p.repository.Delete(ctx, projectServiceID); err != nil {
		return internalProjectServiceError("DELETE_PROJECT_SERVICE_FAILED")
	}

	if apiErr := p.statusSync.Sync(ctx, projectService.ProjectID); apiErr != nil {
		return apiErr
	}

	return nil
}

func projectServiceFromRequest(request dto.ProjectServiceRequest) models.ProjectService {
	return models.ProjectService{
		ProjectID:      request.ProjectID,
		Name:           strings.TrimSpace(request.Name),
		Type:           request.Type,
		URL:            strings.TrimSpace(request.URL),
		RepoURL:        strings.TrimSpace(request.RepoURL),
		Status:         request.Status,
		HealthCheckURL: strings.TrimSpace(request.HealthCheckURL),
	}
}

func projectServiceFromCreateRequest(request dto.ProjectServiceCreateRequest) models.ProjectService {
	return models.ProjectService{
		ProjectID:      request.ProjectID,
		Name:           strings.TrimSpace(request.Name),
		Type:           request.Type,
		URL:            strings.TrimSpace(request.URL),
		RepoURL:        strings.TrimSpace(request.RepoURL),
		HealthCheckURL: strings.TrimSpace(request.HealthCheckURL),
	}
}

func parseProjectServiceID(id string) (uuid.UUID, errors.ApiError) {
	projectServiceID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("PROJECT_SERVICE_ID_INVALID"),
		)
	}

	return projectServiceID, nil
}

func projectServiceRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("PROJECT_SERVICE_NOT_FOUND"),
		)
	}

	return internalProjectServiceError(detail)
}

func internalProjectServiceError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}
