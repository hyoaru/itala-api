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
