package app

func (a *App) RegisterRoutes() {
	api := a.Router.Group("/api/v1")

	projects := api.Group("/projects")
	{
		projects.GET("", a.ProjectsHandler.List())
		projects.POST("", a.ProjectsHandler.Create())
	}

	auth := api.Group("/auth")
	{
		auth.POST("/login", a.AuthHandler.Login())
	}
}
