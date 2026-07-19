package app

import (
	"net/http"

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
	api.GET("/healthcheck", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	projects := api.Group("/projects")
	projects.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		projects.GET("", a.ProjectsHandler.List())
		projects.GET("/:id", a.ProjectsHandler.FindByID())
		projects.POST("", a.ProjectsHandler.Create())
		projects.PUT("/:id", a.ProjectsHandler.Update())
		projects.DELETE("/:id", a.ProjectsHandler.Delete())
	}

	customers := api.Group("/customers")
	customers.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		customers.GET("", a.CustomersHandler.List())
		customers.GET("/:id", a.CustomersHandler.FindByID())
		customers.POST("", a.CustomersHandler.Create())
		customers.PUT("/:id", a.CustomersHandler.Update())
		customers.DELETE("/:id", a.CustomersHandler.Delete())
	}

	customerProjects := api.Group("/customer-projects")
	customerProjects.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		customerProjects.GET("", a.CustomerProjectsHandler.List())
		customerProjects.GET("/:id", a.CustomerProjectsHandler.FindByID())
		customerProjects.POST("", a.CustomerProjectsHandler.Create())
		customerProjects.PUT("/:id", a.CustomerProjectsHandler.Update())
		customerProjects.DELETE("/:id", a.CustomerProjectsHandler.Delete())
	}

	projectServices := api.Group("/project-services")
	projectServices.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		projectServices.GET("", a.ProjectServicesHandler.List())
		projectServices.GET("/:id", a.ProjectServicesHandler.FindByID())
		projectServices.POST("", a.ProjectServicesHandler.Create())
		projectServices.PUT("/:id", a.ProjectServicesHandler.Update())
		projectServices.DELETE("/:id", a.ProjectServicesHandler.Delete())
	}

	serviceChecks := api.Group("/service-checks")
	serviceChecks.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		serviceChecks.GET("", a.ServiceChecksHandler.List())
		serviceChecks.GET("/:id", a.ServiceChecksHandler.FindByID())
	}

	serviceLogs := api.Group("/service-logs")
	serviceLogs.Use(middlewares.AuthRequired(a.AuthService), a.csrfMiddleware())
	{
		serviceLogs.GET("", a.ServiceLogsHandler.List())
		serviceLogs.GET("/:id", a.ServiceLogsHandler.FindByID())
		serviceLogs.POST("", a.ServiceLogsHandler.Create())
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
