package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type IdempotencyAccountRepository struct {
	inner port.AccountRepository
	store idempotency.IdempotencyStore
}

func NewIdempotencyAccountRepository(inner port.AccountRepository, store idempotency.IdempotencyStore) *IdempotencyAccountRepository {
	return &IdempotencyAccountRepository{inner: inner, store: store}
}

func (r *IdempotencyAccountRepository) AdjustBalance(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
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

	if err = r.inner.AdjustBalance(ctx, userID, accountID, idempotencyKey, delta); err != nil {
		_ = r.store.Release(ctx, lock)
		return err
	}

	_ = r.store.Commit(ctx, lock, "null")
	return nil
}

func (r *IdempotencyAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	return r.inner.Create(ctx, userID, account)
}

func (r *IdempotencyAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	return r.inner.Find(ctx, userID, query)
}

func (r *IdempotencyAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	return r.inner.FindOne(ctx, userID, id)
}

func (r *IdempotencyAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	return r.inner.Update(ctx, userID, account)
}

func (r *IdempotencyAccountRepository) Archive(ctx context.Context, userID string, id string) error {
	return r.inner.Archive(ctx, userID, id)
}

func (r *IdempotencyAccountRepository) Restore(ctx context.Context, userID string, id string) error {
	return r.inner.Restore(ctx, userID, id)
}
