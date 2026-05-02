package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestProjectHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("list returns empty project data", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler()
		router.GET("/projects", handler.List())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/projects", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != "{\"data\":[]}" {
			t.Fatalf("Expected empty data response, got %q", w.Body.String())
		}
	})

	t.Run("create returns success", func(t *testing.T) {
		router := gin.New()
		handler := NewProjectHandler()
		router.POST("/projects", handler.Create())
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/projects", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != "\"success\"" {
			t.Fatalf("Expected success response, got %q", w.Body.String())
		}
	})
}
