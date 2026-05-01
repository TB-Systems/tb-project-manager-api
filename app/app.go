package app

import (
	"github.com/TB-Systems/tb-project-manager-api/handlers"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Router          *gin.Engine
	ProjectsHandler *handlers.Project
}

func NewApp(router *gin.Engine, pool *pgxpool.Pool) *App {
	return &App{
		Router:          router,
		ProjectsHandler: handlers.NewProjectsHandler(),
	}
}
