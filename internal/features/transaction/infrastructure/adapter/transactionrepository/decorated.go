package transaction

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
)

type DecoratedTransactionRepository struct {
	inner port.TransactionRepository
}

func NewDecoratedTransactionRepository(inner port.TransactionRepository) *DecoratedTransactionRepository {
	return &DecoratedTransactionRepository{inner: NewLoggingTransactionRepository(inner)}
}

func (c *DecoratedTransactionRepository) Create(
	ctx context.Context,
	userID string,
	transaction entity.Transaction,
) error {
	return c.inner.Create(ctx, userID, transaction)
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
