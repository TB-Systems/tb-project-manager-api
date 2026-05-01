package main

import (
	"fmt"
	"os"

	"github.com/TB-Systems/tb-project-manager-api/app"
	"github.com/TB-Systems/tb-project-manager-api/database"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	db, err := database.Connect(os.Getenv("DB_CONNECTION_STRING"))
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

	app := app.NewApp(gin.Default(), db)
	app.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		panic("PORT_ENV_NOT_SET")
	}

	fmt.Printf("Starting Server on port :%s\n", port)
	if err := app.Router.Run(":" + port); err != nil {
		panic(err)
	}
}
