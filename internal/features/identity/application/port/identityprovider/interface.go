package identity

import (
	"context"

	entity "github.com/hyoaru/itala-api/internal/features/identity/domain/entity"
)

type IdentityProvider interface {
	ValidateToken(ctx context.Context, token string) (entity.Claims, error)
}
