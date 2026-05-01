package main

import (
	"context"
	"fmt"
	"os"

	"github.com/TB-Systems/tb-project-manager-api/app"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, os.Getenv("DB_CONNECTION_STRING"))
	if err != nil {
		panic(err)
	}

	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		panic(err)
	}

	app := app.NewApp(gin.Default(), pool)
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
