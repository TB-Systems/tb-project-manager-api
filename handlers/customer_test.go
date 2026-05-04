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

func TestCustomerHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list returns paginated customers", func(t *testing.T) {
		router := gin.New()
		handler := NewCustomerHandler(&fakeCustomerService{})
		router.GET("/customers", handler.List())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/customers?page=2&limit=5", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"page":2`) {
			t.Fatalf("Expected paginated response, got %q", w.Body.String())
		}
	})

	t.Run("find by id returns customer", func(t *testing.T) {
		router := gin.New()
		handler := NewCustomerHandler(&fakeCustomerService{})
		router.GET("/customers/:id", handler.FindByID())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/customers/8ec5a83a-f508-45de-b73f-6a96a5b32a8f", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"name":"TB Systems"`) {
			t.Fatalf("Expected customer response, got %q", w.Body.String())
		}
	})

	t.Run("create decodes and returns created customer", func(t *testing.T) {
		router := gin.New()
		handler := NewCustomerHandler(&fakeCustomerService{})
		router.POST("/customers", handler.Create())
		w := httptest.NewRecorder()
		body := `{"name":"TB Systems","slug":"tb-systems","document":"04.252.011/0001-10","document_type":2,"email":"contact@tbsystems.com.br","status":1}`
		req := httptest.NewRequest(http.MethodPost, "/customers", strings.NewReader(body))

		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		if !strings.Contains(w.Body.String(), `"slug":"tb-systems"`) {
			t.Fatalf("Expected created customer response, got %q", w.Body.String())
		}
	})

	t.Run("delete returns success", func(t *testing.T) {
		router := gin.New()
		handler := NewCustomerHandler(&fakeCustomerService{})
		router.DELETE("/customers/:id", handler.Delete())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodDelete, "/customers/8ec5a83a-f508-45de-b73f-6a96a5b32a8f", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != `{"message":"success"}` {
			t.Fatalf("Expected success response, got %q", w.Body.String())
		}
	})
}

type fakeCustomerService struct{}

func (f *fakeCustomerService) List(context.Context, commonsmodels.PaginatedParams) (commonsmodels.PaginatedResponse[dto.CustomerResponse], commonsErrors.ApiError) {
	return commonsmodels.PaginatedResponse[dto.CustomerResponse]{
		Items:     []dto.CustomerResponse{{Name: "TB Systems", Slug: "tb-systems"}},
		PageCount: 3,
		Page:      2,
	}, nil
}

func (f *fakeCustomerService) FindByID(context.Context, string) (dto.CustomerResponse, commonsErrors.ApiError) {
	return dto.CustomerResponse{Name: "TB Systems", Slug: "tb-systems"}, nil
}

func (f *fakeCustomerService) Create(_ context.Context, request dto.CustomerRequest) (dto.CustomerResponse, commonsErrors.ApiError) {
	return dto.CustomerResponse{
		Name:         request.Name,
		Slug:         request.Slug,
		Document:     request.Document,
		DocumentType: models.CustomerDocumentTypeCNPJ,
		Email:        request.Email,
		Status:       models.CustomerStatusActive,
	}, nil
}

func (f *fakeCustomerService) Update(_ context.Context, _ string, request dto.CustomerRequest) (dto.CustomerResponse, commonsErrors.ApiError) {
	return dto.CustomerResponse{
		Name:         request.Name,
		Slug:         request.Slug,
		Document:     request.Document,
		DocumentType: request.DocumentType,
		Email:        request.Email,
		Status:       request.Status,
	}, nil
}

func (f *fakeCustomerService) Delete(context.Context, string) commonsErrors.ApiError {
	return nil
}
