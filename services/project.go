package services

import (
	"context"
	stderrors "errors"
	"math"
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

type Project interface {
	List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ProjectResponse], errors.ApiError)
	Overview(ctx context.Context) (commonsmodels.ResponseList[dto.ProjectOverviewResponse], errors.ApiError)
	FindByID(ctx context.Context, id string) (dto.ProjectResponse, errors.ApiError)
	Create(ctx context.Context, request dto.ProjectRequest) (dto.ProjectResponse, errors.ApiError)
	Update(ctx context.Context, id string, request dto.ProjectRequest) (dto.ProjectResponse, errors.ApiError)
	Delete(ctx context.Context, id string) errors.ApiError
}

type project struct {
	repository  repositories.Project
	billingSync CustomerProjectBillingSync
}

func NewProjectService(repository repositories.Project, billingSync ...CustomerProjectBillingSync) Project {
	service := project{repository: repository}
	if len(billingSync) > 0 {
		service.billingSync = billingSync[0]
	}

	return service
}

func (p project) List(ctx context.Context, params commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ProjectResponse], errors.ApiError) {
	if apiErr := p.syncBilling(ctx); apiErr != nil {
		return commonsmodels.PaginatedResponse[dto.ProjectResponse]{}, apiErr
	}

	projects, total, err := p.repository.List(ctx, params)
	if err != nil {
		return commonsmodels.PaginatedResponse[dto.ProjectResponse]{}, internalProjectError("LIST_PROJECTS_FAILED")
	}

	items := make([]dto.ProjectResponse, 0, len(projects))
	for _, project := range projects {
		items = append(items, dto.ProjectResponseFromModel(project))
	}

	return commonsmodels.PaginatedResponse[dto.ProjectResponse]{
		Items:     items,
		PageCount: pageCount(total, params.Limit),
		Page:      int64(params.Page),
	}, nil
}

func (p project) Overview(ctx context.Context) (commonsmodels.ResponseList[dto.ProjectOverviewResponse], errors.ApiError) {
	if apiErr := p.syncBilling(ctx); apiErr != nil {
		return commonsmodels.ResponseList[dto.ProjectOverviewResponse]{}, apiErr
	}

	projects, err := p.repository.Overview(ctx)
	if err != nil {
		return commonsmodels.ResponseList[dto.ProjectOverviewResponse]{}, internalProjectError("LIST_PROJECTS_OVERVIEW_FAILED")
	}

	items := make([]dto.ProjectOverviewResponse, 0, len(projects))
	for _, project := range projects {
		items = append(items, dto.ProjectOverviewResponseFromModel(project))
	}

	return commonsmodels.ResponseList[dto.ProjectOverviewResponse]{
		Items: items,
		Total: len(items),
	}, nil
}

func (p project) FindByID(ctx context.Context, id string) (dto.ProjectResponse, errors.ApiError) {
	projectID, apiErr := parseProjectID(id)
	if apiErr != nil {
		return dto.ProjectResponse{}, apiErr
	}

	if apiErr := p.syncBilling(ctx); apiErr != nil {
		return dto.ProjectResponse{}, apiErr
	}

	project, err := p.repository.FindByID(ctx, projectID)
	if err != nil {
		return dto.ProjectResponse{}, projectRepositoryError(err, "FIND_PROJECT_FAILED")
	}

	return dto.ProjectResponseFromModel(project), nil
}

func (p project) syncBilling(ctx context.Context) errors.ApiError {
	if p.billingSync == nil {
		return nil
	}

	return p.billingSync.SyncOverdue(ctx)
}

func (p project) Create(ctx context.Context, request dto.ProjectRequest) (dto.ProjectResponse, errors.ApiError) {
	project := projectFromRequest(request)
	project.Status = models.ProjectStatusBacklog

	exists, err := p.repository.SlugExists(ctx, project.Slug, nil)
	if err != nil {
		return dto.ProjectResponse{}, internalProjectError("CHECK_PROJECT_SLUG_FAILED")
	}
	if exists {
		return dto.ProjectResponse{}, errors.NewApiError(
			http.StatusConflict,
			errors.BadRequestError("PROJECT_SLUG_ALREADY_EXISTS"),
		)
	}

	createdProject, err := p.repository.Create(ctx, project)
	if err != nil {
		return dto.ProjectResponse{}, internalProjectError("CREATE_PROJECT_FAILED")
	}

	return dto.ProjectResponseFromModel(createdProject), nil
}

func (p project) Update(ctx context.Context, id string, request dto.ProjectRequest) (dto.ProjectResponse, errors.ApiError) {
	projectID, apiErr := parseProjectID(id)
	if apiErr != nil {
		return dto.ProjectResponse{}, apiErr
	}

	currentProject, err := p.repository.FindByID(ctx, projectID)
	if err != nil {
		return dto.ProjectResponse{}, projectRepositoryError(err, "FIND_PROJECT_FAILED")
	}

	project := projectFromRequest(request)
	project.ID = projectID
	project.Status = currentProject.Status

	exists, err := p.repository.SlugExists(ctx, project.Slug, &projectID)
	if err != nil {
		return dto.ProjectResponse{}, internalProjectError("CHECK_PROJECT_SLUG_FAILED")
	}
	if exists {
		return dto.ProjectResponse{}, errors.NewApiError(
			http.StatusConflict,
			errors.BadRequestError("PROJECT_SLUG_ALREADY_EXISTS"),
		)
	}

	updatedProject, err := p.repository.Update(ctx, project)
	if err != nil {
		return dto.ProjectResponse{}, internalProjectError("UPDATE_PROJECT_FAILED")
	}

	return dto.ProjectResponseFromModel(updatedProject), nil
}

func (p project) Delete(ctx context.Context, id string) errors.ApiError {
	projectID, apiErr := parseProjectID(id)
	if apiErr != nil {
		return apiErr
	}

	if _, err := p.repository.FindByID(ctx, projectID); err != nil {
		return projectRepositoryError(err, "FIND_PROJECT_FAILED")
	}

	if err := p.repository.Delete(ctx, projectID); err != nil {
		return internalProjectError("DELETE_PROJECT_FAILED")
	}

	return nil
}

func projectFromRequest(request dto.ProjectRequest) models.Project {
	return models.Project{
		Name:        strings.TrimSpace(request.Name),
		Description: strings.TrimSpace(request.Description),
		Slug:        strings.TrimSpace(request.Slug),
		RepoURL:     strings.TrimSpace(request.RepoURL),
	}
}

func parseProjectID(id string) (uuid.UUID, errors.ApiError) {
	projectID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, errors.NewApiError(
			http.StatusBadRequest,
			errors.BadRequestError("PROJECT_ID_INVALID"),
		)
	}

	return projectID, nil
}

func projectRepositoryError(err error, detail string) errors.ApiError {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewApiError(
			http.StatusNotFound,
			errors.NotFoundError("PROJECT_NOT_FOUND"),
		)
	}

	return internalProjectError(detail)
}

func internalProjectError(detail string) errors.ApiError {
	return errors.NewApiError(
		http.StatusInternalServerError,
		errors.InternalServerError(detail),
	)
}

func pageCount(total int64, limit int32) int64 {
	if total == 0 || limit <= 0 {
		return 0
	}

	return int64(math.Ceil(float64(total) / float64(limit)))
}
