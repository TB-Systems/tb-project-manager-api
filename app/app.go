package app

import (
	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/TB-Systems/tb-project-manager-api/handlers"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Router          *gin.Engine
	Config          config.Config
	ProjectsHandler *handlers.Project
	AuthHandler     *handlers.Auth
	AuthService     services.Auth
}

func NewApp(router *gin.Engine, db *gorm.DB, cfg config.Config) *App {
	authRepository := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepository)

	return &App{
		Router:          router,
		Config:          cfg,
		ProjectsHandler: handlers.NewProjectHandler(),
		AuthHandler:     handlers.NewAuthHandler(authService),
		AuthService:     authService,
	}
}
