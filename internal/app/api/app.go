package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	api "github.com/hyoaru/itala-api/internal/app/api/handlers"
	"github.com/hyoaru/itala-api/internal/shared/logger"
)

type Config struct {
	Addr string
}

type App struct {
	Config Config
}

func (app *App) Run() error {
	mux := chi.NewRouter()
	mux.Use(middleware.RequestID)
	mux.Use(middleware.Logger)
	mux.Use(middleware.Recoverer)
	mux.Use(middleware.Timeout(60 * time.Second))

	transactionHandler := &api.TransactionHandler{}

	mux.Post("/transactions", transactionHandler.Create)

	srv := &http.Server{
		Addr:         app.Config.Addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}

	logger.Debug(fmt.Sprintf("Server has started at %s", srv.Addr))
	return srv.ListenAndServe()
}
