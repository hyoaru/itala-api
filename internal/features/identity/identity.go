package identity

import (
	"context"

	identityproviderport "github.com/hyoaru/itala-api/internal/features/identity/application/port/identityprovider"
	entity "github.com/hyoaru/itala-api/internal/features/identity/domain/entity"
	identityprovideradapter "github.com/hyoaru/itala-api/internal/features/identity/infrastructure/adapter/identityprovider"
	identitycontext "github.com/hyoaru/itala-api/internal/features/identity/infrastructure/context"
)

type (
	User   = entity.User
	Claims = entity.Claims
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
	idp := identityprovideradapter.NewCognitoIdentityProvider(region, userPoolID)
	return identityprovideradapter.NewDecoratedIdentityProvider(idp)
}
