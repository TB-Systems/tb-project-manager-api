package main

import (
	"fmt"
	"os"

	"github.com/TB-Systems/tb-project-manager-api/app"
	"github.com/TB-Systems/tb-project-manager-api/config"
	"github.com/TB-Systems/tb-project-manager-api/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	cfg := config.Load()

	db, err := database.Connect(cfg.DBConnectionString)
	if err != nil {
		panic(err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		panic(err)
	}

	if err := database.Migrate(db); err != nil {
		panic(err)
	}

	router := gin.Default()
	if err := router.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		panic(err)
	}

	app := app.NewApp(router, db, cfg)
	app.RegisterRoutes()
	if err := app.StartJobs(); err != nil {
		panic(err)
	}
	defer app.StopJobs()

	port := cfg.Port
	if port == "" {
		panic("PORT_ENV_NOT_SET")
	}

	fmt.Printf("Starting Server on port :%s\n", port)
	if err := app.Router.Run(":" + port); err != nil {
		panic(err)
	}
}
