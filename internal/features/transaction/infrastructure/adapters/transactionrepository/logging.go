package transaction

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingTransactionRepository struct {
	inner port.TransactionRepository
}

func NewLoggingTransactionRepository(inner port.TransactionRepository) *LoggingTransactionRepository {
	return &LoggingTransactionRepository{inner: inner}
}

func (c *LoggingTransactionRepository) Create(ctx context.Context, userID string, transaction entities.Transaction) error {
	logger.Debug("Creating transaction", "amount", transaction.Amount, "type", transaction.Type)

	if err := c.inner.Create(ctx, userID, transaction); err != nil {
		logger.Warn("Failed to create transaction", "error", err)
		return err
	}

	logger.Info("Transaction created", "amount", transaction.Amount, "type", transaction.Type)

	return nil
}
