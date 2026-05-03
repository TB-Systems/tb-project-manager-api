package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/gin-gonic/gin"
)

func TestCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows configured origin with credentials", func(t *testing.T) {
		router := corsRouter(config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://localhost:5173")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
			t.Fatalf("Expected allowed origin header, got %q", got)
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("Expected credentials true, got %q", got)
		}
		if got := w.Header().Get("Vary"); got != "Origin" {
			t.Fatalf("Expected Vary Origin, got %q", got)
		}
	})

	t.Run("does not allow unconfigured origin", func(t *testing.T) {
		router := corsRouter(config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "https://evil.example")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Expected no allowed origin header, got %q", got)
		}
	})

	t.Run("handles preflight for allowed origin", func(t *testing.T) {
		router := corsRouter(config.Config{CORSAllowedOrigins: []string{"http://localhost:5173"}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodOptions, "/", nil)
		req.Header.Set("Origin", "http://localhost:5173")
		req.Header.Set("Access-Control-Request-Method", http.MethodPost)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Fatalf("Expected status %d, got %d", http.StatusNoContent, w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Methods"); got != corsAllowedMethods {
			t.Fatalf("Expected allowed methods %q, got %q", corsAllowedMethods, got)
		}
		if got := w.Header().Get("Access-Control-Allow-Headers"); got != corsAllowedHeaders {
			t.Fatalf("Expected allowed headers %q, got %q", corsAllowedHeaders, got)
		}
	})

	t.Run("ignores wildcard origin for credentialed cors", func(t *testing.T) {
		router := corsRouter(config.Config{CORSAllowedOrigins: []string{"*"}})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Origin", "http://localhost:5173")

		router.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Fatalf("Expected wildcard not to be reflected, got %q", got)
		}
	})
}

func TestAllowedOriginSet(t *testing.T) {
	origins := allowedOriginSet([]string{" http://localhost:5173 ", "", "*"})

	if _, ok := origins["http://localhost:5173"]; !ok {
		t.Fatal("Expected trimmed origin to be allowed")
	}
	if _, ok := origins["*"]; ok {
		t.Fatal("Expected wildcard origin to be ignored")
	}
	if _, ok := origins[""]; ok {
		t.Fatal("Expected blank origin to be ignored")
	}
}

func corsRouter(cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(CORS(cfg))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.OPTIONS("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}
