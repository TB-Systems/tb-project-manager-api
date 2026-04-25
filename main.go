package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

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

	// api := api.NewApi(gin.Default(), pool)

	// api.RegisterRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		panic("PORT_ENV_NOT_SET")
	}

	fmt.Printf("Starting Server on port :%s\n", port)

	// Keep the process running under Air until API routes are wired.
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	if err := http.ListenAndServe(":"+port, nil); err != nil {
		panic(err)
	}
}
