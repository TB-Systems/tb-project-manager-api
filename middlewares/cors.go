package middlewares

import (
	"net/http"
	"strings"

	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/gin-gonic/gin"
)

const (
	corsAllowedMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	corsAllowedHeaders = "Content-Type,X-CSRF-Token"
)

func CORS(cfg config.Config) gin.HandlerFunc {
	allowedOrigins := allowedOriginSet(cfg.CORSAllowedOrigins)

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if _, allowed := allowedOrigins[origin]; allowed {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Methods", corsAllowedMethods)
			c.Header("Access-Control-Allow-Headers", corsAllowedHeaders)
			c.Header("Vary", "Origin")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func allowedOriginSet(origins []string) map[string]struct{} {
	result := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin != "" && origin != "*" {
			result[origin] = struct{}{}
		}
	}

	return result
}
