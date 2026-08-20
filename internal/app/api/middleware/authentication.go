package api

import (
	"errors"
	"net/http"
	"strings"

	"github.com/hyoaru/itala-api/internal/features/identity"
)

func Authentication(idp identity.IdentityProvider) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authorization := r.Header.Get("Authorization")
			scheme, token, ok := strings.Cut(authorization, " ")

			if !ok || scheme != "Bearer" || token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, err := idp.ValidateToken(ctx, token)
			if err != nil {
				if errors.Is(err, identity.ErrTokenExpired) || errors.Is(err, identity.ErrTokenInvalid) {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}

				http.Error(w, "internal server error", http.StatusInternalServerError)
				return
			}

			user := &identity.User{ID: claims.Subject}
			ctxWithUser := identity.WithUser(ctx, user)
			next.ServeHTTP(w, r.WithContext(ctxWithUser))
		}
		return http.HandlerFunc(fn)
	}
}
