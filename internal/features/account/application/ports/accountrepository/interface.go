package account

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
)

type AccountRepository interface {
	Create(ctx context.Context, userID string, account entities.Account) error
}
