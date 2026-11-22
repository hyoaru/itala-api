package ledgerrepository

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingLedgerRepository struct {
	inner port.LedgerRepository
}

func NewLoggingLedgerRepository(inner port.LedgerRepository) *LoggingLedgerRepository {
	return &LoggingLedgerRepository{inner: inner}
}

func (r *LoggingLedgerRepository) RecordTransaction(ctx context.Context, userID string, input port.RecordTransactionInput) error {
	logger.Debug("Recording transaction", "id", input.TransactionID)
	if err := r.inner.RecordTransaction(ctx, userID, input); err != nil {
		logger.Warn("Failed to record transaction", "id", input.TransactionID, "error", err)
		return err
	}
	logger.Info("Transaction recorded", "id", input.TransactionID)
	return nil
}

func (r *LoggingLedgerRepository) FindIdempotency(ctx context.Context, userID string, keyHash string) (port.IdempotencyRecord, bool, error) {
	return r.inner.FindIdempotency(ctx, userID, keyHash)
}

func (r *LoggingLedgerRepository) FindTransactionDetail(ctx context.Context, userID string, occurredAt string, id string) (entity.Transaction, []entity.Posting, error) {
	return r.inner.FindTransactionDetail(ctx, userID, occurredAt, id)
}
