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

func (r *IdempotencyTransactionRepository) Find(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error) {
	return r.inner.Find(ctx, userID, query)
}

func (r *IdempotencyTransactionRepository) FindOne(ctx context.Context, userID string, id string) (entity.Transaction, error) {
	return r.inner.FindOne(ctx, userID, id)
}

func (r *IdempotencyTransactionRepository) Update(ctx context.Context, userID string, transaction entity.Transaction) error {
	return r.inner.Update(ctx, userID, transaction)
}

func (r *IdempotencyTransactionRepository) Delete(ctx context.Context, userID string, id string) error {
	return r.inner.Delete(ctx, userID, id)
}
