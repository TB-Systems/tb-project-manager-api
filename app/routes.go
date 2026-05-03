package app

import (
	"github.com/TB-Systems/go-commons/utils"
	"github.com/TB-Systems/tb-project-manager-api/constants"
	"github.com/TB-Systems/tb-project-manager-api/middlewares"
	"github.com/gin-gonic/gin"
)

func (a *App) RegisterRoutes() {
	a.Router.Use(
		utils.CORS(utils.CORSConfig{
			AllowedOrigins:   a.Config.CORSAllowedOrigins,
			AllowCredentials: true,
		}),
		utils.SecurityHeaders(utils.SecurityHeadersConfig{
			Production: a.Config.IsProduction(),
		}),
		utils.RequireHTTPS(utils.HTTPSConfig{
			Production:     a.Config.IsProduction(),
			TrustedProxies: a.Config.TrustedProxies,
		}),
	)

	api := a.Router.Group("/api/v1")

	projects := api.Group("/projects")
	projects.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		projects.GET("", a.ProjectsHandler.List())
		projects.POST("", a.ProjectsHandler.Create())
	}

	auth := api.Group("/auth")
	{
		auth.POST("/login", a.AuthHandler.Login())
		auth.POST("/logout", middlewares.AuthRequired(a.AuthService), a.csrfMiddleware(), a.AuthHandler.Logout())
	}
}

func (a *App) csrfMiddleware() gin.HandlerFunc {
	return utils.CSRFRequired(utils.CSRFConfig{
		SessionCookieName: constants.SessionCookieName,
		HeaderName:        constants.CSRFHeaderName,
		Validate:          a.AuthService.ValidateCSRF,
	})
}
