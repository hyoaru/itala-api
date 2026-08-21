package transaction

import (
	"context"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
	"github.com/shopspring/decimal"
)

type LoggingTransactionRepository struct {
	inner port.TransactionRepository
}

func NewLoggingTransactionRepository(inner port.TransactionRepository) *LoggingTransactionRepository {
	return &LoggingTransactionRepository{inner: inner}
}

func (c *LoggingTransactionRepository) Create(
	ctx context.Context,
	userID string,
	amount decimal.Decimal,
	transactionType valueobjects.TransactionType,
	categoryID string,
	description string,
	occurredAt time.Time,
) error {
	logger.Debug("Creating transaction", "amount", amount, "type", transactionType)

	if err := c.inner.Create(
		ctx,
		userID,
		amount,
		transactionType,
		categoryID,
		description,
		occurredAt,
	); err != nil {
		logger.Warn("Failed to create transaction", "error", err)
		return err
	}

	logger.Info("Transaction created", "amount", amount, "type", transactionType)

	return nil
}
