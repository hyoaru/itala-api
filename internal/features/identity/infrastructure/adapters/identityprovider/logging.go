package identity

import (
	identityprovider "github.com/hyoaru/itala-api/internal/features/identity/application/ports/identityprovider"
	entities "github.com/hyoaru/itala-api/internal/features/identity/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingIdentityProvider struct {
	inner identityprovider.IdentityProvider
}

func NewLoggingIdentityProvider(inner identityprovider.IdentityProvider) *LoggingIdentityProvider {
	return &LoggingIdentityProvider{inner: inner}
}

func (idp *LoggingIdentityProvider) ValidateToken(token string) (entities.Claims, error) {
	result, err := idp.inner.ValidateToken(token)
	if err != nil {
		logger.Warn("Access token validation failed")
		return result, err
	}

	logger.Debug(
		"Access token validation successful",
		"user_id", result.Subject,
	)

	return result, err
}
