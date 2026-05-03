package middlewares

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/gin-gonic/gin"
)

func TestSecurityHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("sets common security headers", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeaders(config.Config{AppEnv: config.EnvironmentDevelopment}))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		router.ServeHTTP(w, req)

		if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
			t.Fatalf("Expected nosniff header, got %q", got)
		}
		if got := w.Header().Get("X-Frame-Options"); got != "DENY" {
			t.Fatalf("Expected DENY frame header, got %q", got)
		}
		if got := w.Header().Get("Referrer-Policy"); got != "no-referrer" {
			t.Fatalf("Expected no-referrer policy, got %q", got)
		}
		if got := w.Header().Get("Strict-Transport-Security"); got != "" {
			t.Fatalf("Expected no HSTS in development, got %q", got)
		}
	})

	t.Run("sets hsts in production for https request", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeaders(config.Config{AppEnv: config.EnvironmentProduction}))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.TLS = &tls.ConnectionState{}

		router.ServeHTTP(w, req)

		if got := w.Header().Get("Strict-Transport-Security"); got != "max-age=31536000" {
			t.Fatalf("Expected HSTS header, got %q", got)
		}
	})

	t.Run("sets hsts in production for trusted forwarded https request", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeaders(config.Config{
			AppEnv:         config.EnvironmentProduction,
			TrustedProxies: []string{"192.0.2.1"},
		}))
		router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		req.Header.Set("X-Forwarded-Proto", "https")

		router.ServeHTTP(w, req)

		if got := w.Header().Get("Strict-Transport-Security"); got != "max-age=31536000" {
			t.Fatalf("Expected HSTS header, got %q", got)
		}
	})
}

func TestRequireHTTPS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("allows http in development", func(t *testing.T) {
		router := requireHTTPSRouter(config.Config{AppEnv: config.EnvironmentDevelopment})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("rejects http in production", func(t *testing.T) {
		router := requireHTTPSRouter(config.Config{AppEnv: config.EnvironmentProduction})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUpgradeRequired {
			t.Fatalf("Expected status %d, got %d", http.StatusUpgradeRequired, w.Code)
		}
	})

	t.Run("allows forwarded https in production", func(t *testing.T) {
		router := requireHTTPSRouter(config.Config{
			AppEnv:         config.EnvironmentProduction,
			TrustedProxies: []string{"192.0.2.1"},
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "192.0.2.1:1234"
		req.Header.Set("X-Forwarded-Proto", "HTTPS")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})

	t.Run("rejects spoofed forwarded https from untrusted client", func(t *testing.T) {
		router := requireHTTPSRouter(config.Config{
			AppEnv:         config.EnvironmentProduction,
			TrustedProxies: []string{"127.0.0.1"},
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		req.Header.Set("X-Forwarded-Proto", "https")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusUpgradeRequired {
			t.Fatalf("Expected status %d, got %d", http.StatusUpgradeRequired, w.Code)
		}
	})

	t.Run("allows forwarded https from trusted CIDR", func(t *testing.T) {
		router := requireHTTPSRouter(config.Config{
			AppEnv:         config.EnvironmentProduction,
			TrustedProxies: []string{"10.0.0.0/8"},
		})
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "10.1.2.3:1234"
		req.Header.Set("X-Forwarded-Proto", "https")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestIsTrustedProxy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		remoteAddr     string
		trustedProxies []string
		expected       bool
	}{
		{
			name:           "exact trusted proxy",
			remoteAddr:     "127.0.0.1:1234",
			trustedProxies: []string{"127.0.0.1"},
			expected:       true,
		},
		{
			name:           "trusted cidr",
			remoteAddr:     "10.1.2.3:1234",
			trustedProxies: []string{"10.0.0.0/8"},
			expected:       true,
		},
		{
			name:           "untrusted proxy",
			remoteAddr:     "203.0.113.10:1234",
			trustedProxies: []string{"127.0.0.1"},
			expected:       false,
		},
		{
			name:           "invalid remote address",
			remoteAddr:     "not-an-ip",
			trustedProxies: []string{"127.0.0.1"},
			expected:       false,
		},
		{
			name:           "ignores invalid trusted proxy values",
			remoteAddr:     "127.0.0.1:1234",
			trustedProxies: []string{"", "invalid-cidr/33"},
			expected:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			ctx.Request.RemoteAddr = tt.remoteAddr

			got := isTrustedProxy(ctx, config.Config{TrustedProxies: tt.trustedProxies})
			if got != tt.expected {
				t.Fatalf("Expected trusted proxy to be %v, got %v", tt.expected, got)
			}
		})
	}
}

func requireHTTPSRouter(cfg config.Config) *gin.Engine {
	router := gin.New()
	router.Use(RequireHTTPS(cfg))
	router.GET("/", func(c *gin.Context) { c.Status(http.StatusOK) })
	return router
}
