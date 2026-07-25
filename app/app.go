package app

import (
	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/TB-Systems/tb-project-manager-api/handlers"
	"github.com/TB-Systems/tb-project-manager-api/jobs"
	"github.com/TB-Systems/tb-project-manager-api/repositories"
	"github.com/TB-Systems/tb-project-manager-api/scheduler"
	"github.com/TB-Systems/tb-project-manager-api/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type App struct {
	Router                         *gin.Engine
	Config                         config.Config
	ProjectsHandler                *handlers.Project
	CustomersHandler               *handlers.Customer
	CustomerProjectsHandler        *handlers.CustomerProject
	CustomerProjectInvoicesHandler *handlers.CustomerProjectInvoice
	ProjectServicesHandler         *handlers.ProjectService
	ServiceChecksHandler           *handlers.ServiceCheck
	ServiceLogsHandler             *handlers.ServiceLog
	AuthHandler                    *handlers.Auth
	AuthService                    services.Auth
	Scheduler                      *scheduler.Scheduler
	DailyBillingJob                *jobs.DailyBillingJob
}

func NewApp(router *gin.Engine, db *gorm.DB, cfg config.Config) *App {
	authRepository := repositories.NewAuthRepository(db)
	authService := services.NewAuthService(authRepository)
	customerProjectInvoiceRepository := repositories.NewCustomerProjectInvoiceRepository(db)
	customerStatusRepository := repositories.NewCustomerStatusRepository(db)
	customerStatusSyncService := services.NewCustomerStatusSyncService(customerStatusRepository)
	customerProjectInvoiceService := services.NewCustomerProjectInvoiceService(customerProjectInvoiceRepository, customerStatusSyncService)
	customerProjectBillingSyncService := services.NewCustomerProjectBillingSyncService(customerProjectInvoiceRepository, customerStatusSyncService)
	dailyBillingJob := jobs.NewDailyBillingJob(customerProjectInvoiceRepository, customerProjectBillingSyncService, customerStatusSyncService)
	projectRepository := repositories.NewProjectRepository(db)
	projectService := services.NewProjectService(projectRepository, customerProjectBillingSyncService)
	projectStatusRepository := repositories.NewProjectStatusRepository(db)
	projectStatusSyncService := services.NewProjectStatusSyncService(projectStatusRepository)
	customerRepository := repositories.NewCustomerRepository(db)
	customerService := services.NewCustomerService(customerRepository, customerStatusSyncService)
	customerProjectRepository := repositories.NewCustomerProjectRepository(db)
	customerProjectService := services.NewCustomerProjectService(customerProjectRepository, projectStatusSyncService, customerProjectBillingSyncService, customerStatusSyncService)
	projectServiceRepository := repositories.NewProjectServiceRepository(db)
	projectServiceService := services.NewProjectServiceService(projectServiceRepository, projectStatusSyncService)
	serviceCheckRepository := repositories.NewServiceCheckRepository(db)
	serviceCheckService := services.NewServiceCheckService(serviceCheckRepository)
	serviceLogRepository := repositories.NewServiceLogRepository(db)
	serviceLogService := services.NewServiceLogService(serviceLogRepository)

	return &App{
		Router:                         router,
		Config:                         cfg,
		ProjectsHandler:                handlers.NewProjectHandler(projectService),
		CustomersHandler:               handlers.NewCustomerHandler(customerService),
		CustomerProjectsHandler:        handlers.NewCustomerProjectHandler(customerProjectService),
		CustomerProjectInvoicesHandler: handlers.NewCustomerProjectInvoiceHandler(customerProjectInvoiceService),
		ProjectServicesHandler:         handlers.NewProjectServiceHandler(projectServiceService),
		ServiceChecksHandler:           handlers.NewServiceCheckHandler(serviceCheckService),
		ServiceLogsHandler:             handlers.NewServiceLogHandler(serviceLogService),
		AuthHandler:                    handlers.NewAuthHandler(authService),
		AuthService:                    authService,
		Scheduler:                      scheduler.New(),
		DailyBillingJob:                dailyBillingJob,
	}
}

func (a *App) StartJobs() error {
	if !a.Config.JobsEnabled {
		return nil
	}

	return a.Scheduler.StartDaily("daily-billing", a.Config.DailyBillingJobTime, a.DailyBillingJob.Run)
}

func (a *App) StopJobs() {
	if a.Scheduler != nil {
		a.Scheduler.Stop()
	}
}
