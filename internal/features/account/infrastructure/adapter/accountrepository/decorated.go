package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type DecoratedAccountRepository struct {
	inner port.AccountRepository
}

func NewDecoratedAccountRepository(inner port.AccountRepository) *DecoratedAccountRepository {
	return &DecoratedAccountRepository{inner: NewLoggingAccountRepository(inner)}
}

func (c *DecoratedAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	return c.inner.Create(ctx, userID, account)
}

func (c *DecoratedAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	return c.inner.Find(ctx, userID, query)
}

func (c *DecoratedAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	return c.inner.FindOne(ctx, userID, id)
}

func (c *DecoratedAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	return c.inner.Update(ctx, userID, account)
}
