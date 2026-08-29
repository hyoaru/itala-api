package transaction

import (
	"context"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type DecoratedTransactionRepository struct {
	inner port.TransactionRepository
}

func NewDecoratedTransactionRepository(inner port.TransactionRepository, idempotencyStore idempotency.IdempotencyStore) port.TransactionRepository {
	logging := NewLoggingTransactionRepository(inner)
	idempotency := NewIdempotencyTransactionRepository(logging, idempotencyStore)
	retry := NewRetryTransactionRepository(idempotency, 5, 100*time.Millisecond, 2*time.Second)
	return &DecoratedTransactionRepository{inner: retry}
}

func (c *DecoratedTransactionRepository) Create(
	ctx context.Context,
	userID string,
	transaction entity.Transaction,
	idempotencyKey string,
) error {
	return c.inner.Create(ctx, userID, transaction, idempotencyKey)
}

func (c *DecoratedTransactionRepository) Find(
	ctx context.Context,
	userID string,
	query port.TransactionQuery,
) (port.TransactionPage, error) {
	return c.inner.Find(ctx, userID, query)
}

func (c *DecoratedTransactionRepository) FindOne(ctx context.Context, userID string, id string) (entity.Transaction, error) {
	return c.inner.FindOne(ctx, userID, id)
}

func (c *DecoratedTransactionRepository) Update(ctx context.Context, userID string, transaction entity.Transaction) error {
	return c.inner.Update(ctx, userID, transaction)
}

func (c *DecoratedTransactionRepository) Delete(ctx context.Context, userID string, id string) error {
	return c.inner.Delete(ctx, userID, id)
}
