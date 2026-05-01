package app

import (
	"github.com/TB-Systems/tb-project-manager-api/handlers"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Router          *gin.Engine
	ProjectsHandler *handlers.Project
	AuthHandler     *handlers.Auth
}

func NewApp(router *gin.Engine, db *gorm.DB) *App {
	return &App{
		Router:          router,
		ProjectsHandler: handlers.NewProjectHandler(),
		AuthHandler:     createAuth(db),
	}
}

func createAuth(db *gorm.DB) *handlers.Auth {
	repository := repositories.NewAuthRepository(db)
	service := services.NewAuthService(repository)
	return handlers.NewAuthHandler(service)
}
