package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/TB-Systems/go-commons/commonsmodels"
	commonsErrors "github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/tb-project-manager-api/dto"
	"github.com/TB-Systems/tb-project-manager-api/models"
	"github.com/gin-gonic/gin"
)

func TestProjectHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list returns paginated projects", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler(&fakeProjectService{})
		router.GET("/projects", handler.List())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/projects?page=2&limit=5", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"page":2`) {
			t.Fatalf("Expected paginated response, got %q", w.Body.String())
		}
	})

	t.Run("find by id returns project", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler(&fakeProjectService{})
		router.GET("/projects/:id", handler.FindByID())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/projects/8ec5a83a-f508-45de-b73f-6a96a5b32a8f", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"name":"TB Manager"`) {
			t.Fatalf("Expected project response, got %q", w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"customer_projects"`) {
			t.Fatalf("Expected project detail with customer_projects field, got %q", w.Body.String())
		}
	})

	t.Run("overview returns paginated dashboard projects", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler(&fakeProjectService{})
		router.GET("/dashboard", handler.Overview())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/dashboard", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"customers"`) {
			t.Fatalf("Expected overview response with customers field, got %q", w.Body.String())
		}
		if !strings.Contains(w.Body.String(), `"services"`) {
			t.Fatalf("Expected overview response with services field, got %q", w.Body.String())
		}
	})

	t.Run("create decodes and returns created project", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler(&fakeProjectService{})
		router.POST("/projects", handler.Create())
		w := httptest.NewRecorder()
		body := `{"name":"TB Manager","description":"Internal manager","slug":"tb-manager","repo_url":"https://github.com/TB-Systems/tb-manager"}`
		req := httptest.NewRequest(http.MethodPost, "/projects", strings.NewReader(body))

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"slug":"tb-manager"`) {
			t.Fatalf("Expected created project response, got %q", w.Body.String())
		}
	})

	t.Run("delete returns success", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler(&fakeProjectService{})
		router.DELETE("/projects/:id", handler.Delete())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/projects/8ec5a83a-f508-45de-b73f-6a96a5b32a8f", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != `{"message":"success"}` {
			t.Fatalf("Expected success response, got %q", w.Body.String())
		}
	})
}

type fakeProjectService struct{}

func (f *fakeProjectService) List(context.Context, commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.ProjectResponse], commonsErrors.ApiError) {
	return commonsmodels.PaginatedResponse[dto.ProjectResponse]{
		Items:     []dto.ProjectResponse{{Name: "TB Manager", Slug: "tb-manager"}},
		PageCount: 3,
		Page:      2,
	}, nil
}

func (f *fakeProjectService) Overview(context.Context) (commonsmodels.ResponseList[dto.ProjectOverviewResponse], commonsErrors.ApiError) {
	return commonsmodels.ResponseList[dto.ProjectOverviewResponse]{
		Items: []dto.ProjectOverviewResponse{
			{
				Name:      "TB Manager",
				Slug:      "tb-manager",
				Customers: nil,
				Services:  []dto.ProjectOverviewServiceResponse{{Name: "API"}},
			},
		},
		Total: 1,
	}, nil
}

func (f *fakeProjectService) FindByID(context.Context, string) (dto.ProjectResponse, commonsErrors.ApiError) {
	return dto.ProjectResponse{Name: "TB Manager", Slug: "tb-manager", CustomerProjects: nil}, nil
}

func (f *fakeProjectService) Create(_ context.Context, request dto.ProjectRequest) (dto.ProjectResponse, commonsErrors.ApiError) {
	return dto.ProjectResponse{
		Name:        request.Name,
		Description: request.Description,
		Slug:        request.Slug,
		RepoURL:     request.RepoURL,
		Status:      models.ProjectStatusBacklog,
	}, nil
}

func (f *fakeProjectService) Update(_ context.Context, _ string, request dto.ProjectRequest) (dto.ProjectResponse, commonsErrors.ApiError) {
	return dto.ProjectResponse{
		Name:        request.Name,
		Description: request.Description,
		Slug:        request.Slug,
		RepoURL:     request.RepoURL,
		Status:      models.ProjectStatusBacklog,
	}, nil
}

func (f *fakeProjectService) Delete(context.Context, string) commonsErrors.ApiError {
	return nil
}
