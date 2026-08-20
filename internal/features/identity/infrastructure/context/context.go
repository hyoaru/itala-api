package identity

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/identity/domain/entities"
)

func WithUser(ctx context.Context, user *entities.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

func UserFromContext(ctx context.Context) *entities.User {
	user, _ := ctx.Value(userContextKey).(*entities.User)
	return user
}
