package transaction

import (
	"context"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/shopspring/decimal"
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
	amount decimal.Decimal,
	transactionType valueobjects.TransactionType,
	categoryID string,
	description string,
	occurredAt time.Time,
) error {
	return c.inner.Create(
		ctx,
		userID,
		amount,
		transactionType,
		categoryID,
		description,
		occurredAt,
	)
}
