package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	commonsErrors "github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/gin-gonic/gin"
)

func TestProjectServiceHandlerListPassesProjectIDFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	projectID := "8ec5a83a-f508-45de-b73f-6a96a5b32a8f"
	fakeService := &fakeProjectServiceService{}
	handler := NewProjectServiceHandler(fakeService)
	router := gin.New()
	router.GET("/project-services", handler.List())
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/project-services?page=1&limit=10&project_id="+projectID, nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if fakeService.receivedFilter.ProjectID != projectID {
		t.Fatalf("Expected project id filter %q, got %q", projectID, fakeService.receivedFilter.ProjectID)
	}
}

type fakeProjectServiceService struct {
	receivedFilter dto.ProjectServiceListFilter
}

func (f *fakeProjectServiceService) List(_ context.Context, _ commonsmodels.PaginatedParams, filter dto.ProjectServiceListFilter) (commonsmodels.PaginatedResponse[dto.ProjectServiceResponse], commonsErrors.ApiError) {
	f.receivedFilter = filter
	return commonsmodels.PaginatedResponse[dto.ProjectServiceResponse]{
		Items:     []dto.ProjectServiceResponse{{Name: "API"}},
		PageCount: 1,
		Page:      1,
	}, nil
}

func (f *fakeProjectServiceService) FindByID(context.Context, string) (dto.ProjectServiceResponse, commonsErrors.ApiError) {
	return dto.ProjectServiceResponse{}, nil
}

func (f *fakeProjectServiceService) Create(_ context.Context, request dto.ProjectServiceCreateRequest) (dto.ProjectServiceResponse, commonsErrors.ApiError) {
	return dto.ProjectServiceResponse{Name: request.Name}, nil
}

func (f *fakeProjectServiceService) Update(_ context.Context, _ string, request dto.ProjectServiceRequest) (dto.ProjectServiceResponse, commonsErrors.ApiError) {
	return dto.ProjectServiceResponse{Name: request.Name}, nil
}

func (f *fakeProjectServiceService) Delete(context.Context, string) commonsErrors.ApiError {
	return nil
}
