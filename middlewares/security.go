package middlewares

import (
	"net"
	"net/http"
	"strings"

	"github.com/TB-Systems/go-commons/errors"
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/gin-gonic/gin"
)

func SecurityHeaders(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("Referrer-Policy", "no-referrer")

		if cfg.IsProduction() && isSecureRequest(c, cfg) {
			c.Header("Strict-Transport-Security", "max-age=31536000")
		}

		c.Next()
	}
}

func RequireHTTPS(cfg config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !cfg.IsProduction() || isSecureRequest(c, cfg) {
			c.Next()
			return
		}

		utils.SendErrorResponse(c, errors.NewApiError(
			http.StatusUpgradeRequired,
			errors.BadRequestError("HTTPS_REQUIRED"),
		))
	}
}

func isSecureRequest(c *gin.Context, cfg config.Config) bool {
	return c.Request.TLS != nil || (strings.EqualFold(c.GetHeader("X-Forwarded-Proto"), "https") && isTrustedProxy(c, cfg))
}

func isTrustedProxy(c *gin.Context, cfg config.Config) bool {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		host = c.Request.RemoteAddr
	}

	remoteIP := net.ParseIP(host)
	if remoteIP == nil {
		return false
	}

	for _, trustedProxy := range cfg.TrustedProxies {
		trustedProxy = strings.TrimSpace(trustedProxy)
		if trustedProxy == "" {
			continue
		}

		if strings.Contains(trustedProxy, "/") {
			_, network, err := net.ParseCIDR(trustedProxy)
			if err == nil && network.Contains(remoteIP) {
				return true
			}
			continue
		}

		trustedIP := net.ParseIP(trustedProxy)
		if trustedIP != nil && trustedIP.Equal(remoteIP) {
			return true
		}
	}

	return false
}
