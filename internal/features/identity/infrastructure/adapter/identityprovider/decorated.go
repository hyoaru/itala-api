package identity

import (
	"context"

	identityprovider "github.com/hyoaru/itala-api/internal/features/identity/application/port/identityprovider"
	entity "github.com/hyoaru/itala-api/internal/features/identity/domain/entity"
)

type DecoratedIdentityProvider struct {
	inner identityprovider.IdentityProvider
}

func NewDecoratedIdentityProvider(inner identityprovider.IdentityProvider) *DecoratedIdentityProvider {
	return &DecoratedIdentityProvider{inner: NewLoggingIdentityProvider(inner)}
}

func (idp *DecoratedIdentityProvider) ValidateToken(ctx context.Context, token string) (entity.Claims, error) {
	return idp.inner.ValidateToken(ctx, token)
}
