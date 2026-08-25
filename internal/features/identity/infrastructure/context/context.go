package identity

import (
	"context"

	entitiy "github.com/hyoaru/itala-api/internal/features/identity/domain/entity"
)

type contextKey string

const userContextKey contextKey = "user"

func WithUser(ctx context.Context, user *entitiy.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) *entitiy.User {
	user, _ := ctx.Value(userContextKey).(*entitiy.User)
	return user
}
