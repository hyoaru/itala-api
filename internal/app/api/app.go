package api

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	app "github.com/hyoaru/itala-api/internal/app"
	handler "github.com/hyoaru/itala-api/internal/app/api/handler"
	middleware "github.com/hyoaru/itala-api/internal/app/api/middleware"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type App struct{ server *http.Server }

func New(addr string) app.Application {
	// Composition Root
	idp := identity.NewCognitoIdentityProvider(os.Getenv("AWS_REGION"), os.Getenv("COGNITO_USER_POOL_ID"))

	mux := chi.NewRouter()
	mux.Use(chiMiddleware.RequestID)
	mux.Use(chiMiddleware.Logger)
	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.Timeout(60 * time.Second))

	transactionHandler := &handler.TransactionHandler{}
	mux.Group(func(r chi.Router) {
		r.Use(middleware.Authentication(idp))
		r.Post("/transactions", transactionHandler.Create)
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		WriteTimeout: time.Second * 30,
		ReadTimeout:  time.Second * 10,
		IdleTimeout:  time.Minute,
	}
	return &App{server: server}
}

func (app *App) Run() error {
	logger.Debug(fmt.Sprintf("Server has started at %s", app.server.Addr))
	return app.server.ListenAndServe()
}
