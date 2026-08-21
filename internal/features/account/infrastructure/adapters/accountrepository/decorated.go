package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
)

type DecoratedAccountRepository struct {
	inner port.AccountRepository
}

func NewDecoratedAccountRepository(inner port.AccountRepository) *DecoratedAccountRepository {
	return &DecoratedAccountRepository{inner: NewLoggingAccountRepository(inner)}
}

func (c *DecoratedAccountRepository) Create(ctx context.Context, userID string, name string) error {
	return c.inner.Create(ctx, userID, name)
}
