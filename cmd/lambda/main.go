package main

import (
	"github.com/akrylysov/algnhsa"
	"github.com/hyoaru/itala-api/internal/app/api"
)

func main() {
	app := api.New()
	algnhsa.ListenAndServe(app.Handler, nil)
}
