package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewApp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	app := NewApp(router, nil)

	if app.Router != router {
		t.Fatal("Expected app to keep provided router")
	}
	if app.ProjectsHandler == nil {
		t.Fatal("Expected projects handler to be initialized")
	}
	if app.AuthHandler == nil {
		t.Fatal("Expected auth handler to be initialized")
	}
	if app.AuthService == nil {
		t.Fatal("Expected auth service to be initialized")
	}
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	app := NewApp(router, nil)
	app.RegisterRoutes()

	routes := router.Routes()
	expectedRoutes := map[string]bool{
		"GET /api/v1/projects":     false,
		"POST /api/v1/projects":    false,
		"POST /api/v1/auth/login":  false,
		"POST /api/v1/auth/logout": false,
	}

	for _, route := range routes {
		key := route.Method + " " + route.Path
		if _, ok := expectedRoutes[key]; ok {
			expectedRoutes[key] = true
		}
	}

	for route, found := range expectedRoutes {
		if !found {
			t.Fatalf("Expected route %s to be registered", route)
		}
	}

	t.Run("protected project route rejects missing session", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}
