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
) http.Handler {
	mux := chi.NewRouter()

	mux.Use(chiMiddleware.RequestID)
	mux.Use(chiMiddleware.Logger)
	mux.Use(chiMiddleware.Recoverer)
	mux.Use(chiMiddleware.Timeout(60 * time.Second))

	mux.Group(func(r chi.Router) {
		r.Use(middleware.Authentication(identityProvider))
		r.Post("/categories", categoryHandler.Create)
	})

	return mux
}
