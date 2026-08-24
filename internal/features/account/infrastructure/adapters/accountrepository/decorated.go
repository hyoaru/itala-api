package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
)

type DecoratedAccountRepository struct {
	inner port.AccountRepository
}

func NewDecoratedAccountRepository(inner port.AccountRepository) *DecoratedAccountRepository {
	return &DecoratedAccountRepository{inner: NewLoggingAccountRepository(inner)}
}

func (c *DecoratedAccountRepository) Create(ctx context.Context, userID string, account entities.Account) error {
	return c.inner.Create(ctx, userID, account)
}

func (c *DecoratedAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	return c.inner.Find(ctx, userID, query)
}
