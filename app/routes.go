package app

import "github.com/TB-Systems/tb-project-manager-api/middlewares"

func (a *App) RegisterRoutes() {
	a.Router.Use(middlewares.SecurityHeaders(a.Config), middlewares.RequireHTTPS(a.Config))

	api := a.Router.Group("/api/v1")

	projects := api.Group("/projects")
	projects.Use(middlewares.AuthRequired(a.AuthService))
	{
		projects.GET("", a.ProjectsHandler.List())
		projects.POST("", a.ProjectsHandler.Create())
	}

	auth := api.Group("/auth")
	{
		auth.POST("/login", a.AuthHandler.Login())
		auth.POST("/logout", middlewares.AuthRequired(a.AuthService), a.AuthHandler.Logout())
	}
}
