package identity

import entities "github.com/hyoaru/itala-api/internal/features/identity/domain/entities"

type IdentityProvider interface {
	ValidateToken(token string) (entities.Claims, error)
}
