package identity

import (
	"context"

	identityprovider "github.com/hyoaru/itala-api/internal/features/identity/application/port/identityprovider"
	entity "github.com/hyoaru/itala-api/internal/features/identity/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingIdentityProvider struct {
	inner identityprovider.IdentityProvider
}

func NewLoggingIdentityProvider(inner identityprovider.IdentityProvider) *LoggingIdentityProvider {
	return &LoggingIdentityProvider{inner: inner}
}

func (idp *LoggingIdentityProvider) ValidateToken(ctx context.Context, token string) (entity.Claims, error) {
	result, err := idp.inner.ValidateToken(ctx, token)
	if err != nil {
		logger.Warn("Access token validation failed")
		return result, err
	}

	logger.Info(
		"Access token validation successful",
		"user_id", result.Subject,
	)

	return result, err
}
