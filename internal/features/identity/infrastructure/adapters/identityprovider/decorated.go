package identity

import (
	"context"

	identityprovider "github.com/hyoaru/itala-api/internal/features/identity/application/ports/identityprovider"
	entities "github.com/hyoaru/itala-api/internal/features/identity/domain/entities"
)

type DecoratedIdentityProvider struct {
	inner identityprovider.IdentityProvider
}

func NewDecoratedIdentityProvider(inner identityprovider.IdentityProvider) *DecoratedIdentityProvider {
	return &DecoratedIdentityProvider{inner: NewLoggingIdentityProvider(inner)}
}

func (idp *DecoratedIdentityProvider) ValidateToken(ctx context.Context, token string) (entities.Claims, error) {
	return idp.inner.ValidateToken(ctx, token)
}
