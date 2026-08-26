package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	handler "github.com/hyoaru/itala-api/internal/app/api/handler"
	middleware "github.com/hyoaru/itala-api/internal/app/api/middleware"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

func NewRouter(
	identityProvider identity.IdentityProvider,
	categoryHandler handler.CategoryHandler,
	accountHandler handler.AccountHandler,
	transactionHandler handler.TransactionHandler,
) http.Handler {
	mux := chi.NewRouter()

	mux.Use(chiMiddleware.RequestID)
	mux.Use(chiMiddleware.Logger)
	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.Timeout(60 * time.Second))

	mux.Group(func(r chi.Router) {
		r.Use(middleware.Authentication(identityProvider))

		r.Route("/categories", func(r chi.Router) {
			r.Post("/", categoryHandler.Create)
			r.Get("/", categoryHandler.List)
			r.Get("/{id}", categoryHandler.Get)
			r.Put("/{id}", categoryHandler.Update)
			r.Post("/{id}/archive", categoryHandler.Archive)
			r.Post("/{id}/restore", categoryHandler.Restore)
		})

		r.Route("/accounts", func(r chi.Router) {
			r.Post("/", accountHandler.Create)
			r.Get("/", accountHandler.List)
			r.Get("/{id}", accountHandler.Get)
			r.Put("/{id}", accountHandler.Update)
			r.Post("/{id}/archive", accountHandler.Archive)
			r.Post("/{id}/restore", accountHandler.Restore)
		})

		r.Route("/transactions", func(r chi.Router) {
			r.Post("/", transactionHandler.Create)
			r.Get("/", transactionHandler.List)
			r.Get("/{id}", transactionHandler.Get)
			r.Put("/{id}", transactionHandler.Update)
		})
	})

	return mux
}
