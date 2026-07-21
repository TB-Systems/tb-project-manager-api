package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/gin-gonic/gin"
)

func TestNewApp(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	cfg := config.Config{AppEnv: config.EnvironmentDevelopment}
	app := NewApp(router, nil, cfg)

	if app.Router != router {
		t.Fatal("Expected app to keep provided router")
	}
	if app.Config.AppEnv != cfg.AppEnv {
		t.Fatal("Expected app to keep provided config")
	}
	if app.ProjectsHandler == nil {
		t.Fatal("Expected projects handler to be initialized")
	}
	if app.CustomersHandler == nil {
		t.Fatal("Expected customers handler to be initialized")
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
	app := NewApp(router, nil, config.Config{AppEnv: config.EnvironmentDevelopment})
	app.RegisterRoutes()

	routes := router.Routes()
	expectedRoutes := map[string]bool{
		"GET /api/v1/projects":         false,
		"GET /api/v1/projects/:id":     false,
		"POST /api/v1/projects":        false,
		"PUT /api/v1/projects/:id":     false,
		"DELETE /api/v1/projects/:id":  false,
		"GET /api/v1/customers":        false,
		"GET /api/v1/customers/:id":    false,
		"POST /api/v1/customers":       false,
		"PUT /api/v1/customers/:id":    false,
		"DELETE /api/v1/customers/:id": false,
		"POST /api/v1/auth/login":      false,
		"GET /api/v1/auth/session":     false,
		"POST /api/v1/auth/logout":     false,
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

	t.Run("protected customer route rejects missing session", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/customers", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("Expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})
}
