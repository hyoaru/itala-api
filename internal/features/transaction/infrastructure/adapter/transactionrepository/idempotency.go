package transaction

import (
	"context"
	"errors"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type IdempotencyTransactionRepository struct {
	inner port.TransactionRepository
	store idempotency.IdempotencyStore
}

func NewIdempotencyTransactionRepository(inner port.TransactionRepository, store idempotency.IdempotencyStore) *IdempotencyTransactionRepository {
	return &IdempotencyTransactionRepository{inner: inner, store: store}
}

func (r *IdempotencyTransactionRepository) Create(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
	lock, status, _, err := r.store.Acquire(ctx, idempotencyKey, 900)
	if err != nil {
		return err
	}

	if status != idempotency.IdempotencyStatusAcquired {
		if status == idempotency.IdempotencyStatusLocked {
			return idempotency.ErrResourceLocked
		}
		return nil
	}

	if err = r.inner.Create(ctx, userID, transaction, idempotencyKey); err != nil {
		if !errors.Is(err, port.ErrTransactionExists) {
			_ = r.store.Release(ctx, lock)
			return err
		}
		return err
	}

	_ = r.store.Commit(ctx, lock, "null")
	return nil
}
