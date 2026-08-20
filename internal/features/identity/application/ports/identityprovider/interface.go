package identity

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/identity/domain/entities"
)

type IdentityProvider interface {
	ValidateToken(ctx context.Context, token string) (entities.Claims, error)
}
