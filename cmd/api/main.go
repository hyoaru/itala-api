package main

import (
	"log"

	"github.com/hyoaru/itala-api/internal/app/api"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("failed to load .env")
	}

	app := api.New(":8080")
	log.Fatal(app.Run())
}
