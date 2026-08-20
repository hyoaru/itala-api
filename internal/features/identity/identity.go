package identity

import (
	"context"

	identityproviderport "github.com/hyoaru/itala-api/internal/features/identity/application/ports/identityprovider"
	entities "github.com/hyoaru/itala-api/internal/features/identity/domain/entities"
	identityprovideradapters "github.com/hyoaru/itala-api/internal/features/identity/infrastructure/adapters/identityprovider"
	identitycontext "github.com/hyoaru/itala-api/internal/features/identity/infrastructure/context"
)

type (
	User   = entities.User
	Claims = entities.Claims
)

func WithUser(ctx context.Context, user *User) context.Context {
	return identitycontext.WithUser(ctx, user)
}

func UserFromContext(ctx context.Context) *User {
	return identitycontext.UserFromContext(ctx)
}

type IdentityProvider = identityproviderport.IdentityProvider

var (
	ErrTokenInvalid = identityproviderport.ErrTokenInvalid
	ErrTokenExpired = identityproviderport.ErrTokenExpired
)

func NewCognitoIdentityProvider(region string, userPoolID string) IdentityProvider {
	idp := identityprovideradapters.NewCognitoIdentityProvider(region, userPoolID)
	return identityprovideradapters.NewDecoratedIdentityProvider(idp)
}
