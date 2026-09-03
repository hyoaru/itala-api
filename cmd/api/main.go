package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/hyoaru/itala-api/internal/app/api"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env")
	}

	app := api.New()

	server := &http.Server{
		Addr:         ":8080",
		Handler:      app.Handler,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	logger.Debug(fmt.Sprintf("Server has started at %s", server.Addr))
	log.Fatal(server.ListenAndServe())
}
