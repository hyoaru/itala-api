package identity

import valueobjects "github.com/hyoaru/itala-api/internal/features/identity/domain/valueobjects"

type IdentityProvider interface {
	ValidateToken(token string) (valueobjects.Claims, error)
}
