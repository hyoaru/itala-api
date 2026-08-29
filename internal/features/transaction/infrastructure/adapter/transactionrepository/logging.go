package transaction

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingTransactionRepository struct {
	inner port.TransactionRepository
}

func NewLoggingTransactionRepository(inner port.TransactionRepository) *LoggingTransactionRepository {
	return &LoggingTransactionRepository{inner: inner}
}

func (r *LoggingTransactionRepository) Create(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
	logger.Debug("Creating transaction", "amount", transaction.Amount, "type", transaction.Type)

	if err := r.inner.Create(ctx, userID, transaction, idempotencyKey); err != nil {
		logger.Warn("Failed to create transaction", "error", err)
		return err
	}

	logger.Info("Transaction created", "amount", transaction.Amount, "type", transaction.Type)

	return nil
}

func (r *LoggingTransactionRepository) Find(
	ctx context.Context,
	userID string,
	query port.TransactionQuery,
) (port.TransactionPage, error) {
	logger.Debug("Finding transactions", "query", query)

	result, err := r.inner.Find(ctx, userID, query)
	if err != nil {
		logger.Warn("Failed to find transactions", "error", err)
		return result, err
	}

	logger.Info("Transactions found", "count", len(result.Transactions))

	return result, nil
}

func (r *LoggingTransactionRepository) FindOne(ctx context.Context, userID string, id string) (entity.Transaction, error) {
	logger.Debug("Finding transaction", "id", id)

	result, err := r.inner.FindOne(ctx, userID, id)
	if err != nil {
		logger.Warn("Failed to find transaction", "error", err)
		return result, err
	}

	logger.Info("Transaction found", "id", id)

	return result, nil
}

func (r *LoggingTransactionRepository) Update(ctx context.Context, userID string, transaction entity.Transaction) error {
	logger.Debug("Updating transaction", "amount", transaction.Amount, "type", transaction.Type)

	if err := r.inner.Update(ctx, userID, transaction); err != nil {
		logger.Warn("Failed to update transaction", "error", err)
		return err
	}

	logger.Info("Transaction updated", "amount", transaction.Amount, "type", transaction.Type)

	return nil
}

func (r *LoggingTransactionRepository) Delete(ctx context.Context, userID string, id string) error {
	logger.Debug("Deleting transaction", "id", id)

	if err := r.inner.Delete(ctx, userID, id); err != nil {
		logger.Warn("Failed to delete transaction", "error", err)
		return err
	}

	logger.Info("Transaction deleted", "id", id)

	return nil
}
