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
	Router                  *gin.Engine
	Config                  config.Config
	ProjectsHandler         *handlers.Project
	CustomersHandler        *handlers.Customer
	CustomerProjectsHandler *handlers.CustomerProject
	ProjectServicesHandler  *handlers.ProjectService
	ServiceChecksHandler    *handlers.ServiceCheck
	AuthHandler             *handlers.Auth
	AuthService             services.Auth
}

func NewApp(router *gin.Engine, db *gorm.DB, cfg config.Config) *App {
	authRepository := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepository)
	projectRepository := repositories.NewProjectRepository(db)
	projectService := services.NewProjectService(projectRepository)
	customerRepository := repositories.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepository)
	customerProjectRepository := repositories.NewCustomerProjectRepository(db)
	customerProjectService := services.NewCustomerProjectService(customerProjectRepository)
	projectServiceRepository := repositories.NewProjectServiceRepository(db)
	projectServiceService := services.NewProjectServiceService(projectServiceRepository)
	serviceCheckRepository := repositories.NewServiceCheckRepository(db)
	serviceCheckService := services.NewServiceCheckService(serviceCheckRepository)

	return &App{
		Router:                  router,
		Config:                  cfg,
		ProjectsHandler:         handlers.NewProjectHandler(projectService),
		CustomersHandler:        handlers.NewCustomerHandler(customerService),
		CustomerProjectsHandler: handlers.NewCustomerProjectHandler(customerProjectService),
		ProjectServicesHandler:  handlers.NewProjectServiceHandler(projectServiceService),
		ServiceChecksHandler:    handlers.NewServiceCheckHandler(serviceCheckService),
		AuthHandler:             handlers.NewAuthHandler(authService),
		AuthService:             authService,
	}
}
