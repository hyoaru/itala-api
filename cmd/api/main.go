package main

import (
	"log"

	"github.com/hyoaru/itala-api/internal/app"
	"github.com/hyoaru/itala-api/internal/app/api"
)

func main() {
	var api app.Application = &api.App{
		Config: api.Config{
			Addr: ":8080",
		},
	}

	log.Fatal(api.Run())
}
