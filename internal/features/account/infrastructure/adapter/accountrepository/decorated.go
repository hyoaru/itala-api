package account

import (
	"context"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type DecoratedAccountRepository struct {
	inner port.AccountRepository
}

func NewDecoratedAccountRepository(inner port.AccountRepository, idempotencyStore idempotency.IdempotencyStore) *DecoratedAccountRepository {
	logging := NewLoggingAccountRepository(inner)
	idempotency := NewIdempotencyAccountRepository(logging, idempotencyStore)
	retry := NewRetryAccountRepository(idempotency, 5, 100*time.Millisecond, 2*time.Second)
	return &DecoratedAccountRepository{inner: retry}
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

func (c *DecoratedAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	return c.inner.Delete(ctx, userID, id)
}

func (c *DecoratedAccountRepository) AdjustBalance(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
	return c.inner.AdjustBalance(ctx, userID, accountID, idempotencyKey, delta)
}
