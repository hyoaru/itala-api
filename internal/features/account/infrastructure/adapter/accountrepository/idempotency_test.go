package account

import (
	"context"
	"errors"
	"testing"

	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type fakeIdempotencyStore struct {
	idempotency.IdempotencyStore
	acquire func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error)
	commit  func(ctx context.Context, lock idempotency.IdempotencyLock, resultJSON string) error
	release func(ctx context.Context, lock idempotency.IdempotencyLock) error
}

func (f *fakeIdempotencyStore) Acquire(
	ctx context.Context,
	key string,
	expiresAt uint16,
) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
	return f.acquire(ctx, key, expiresAt)
}

func (f *fakeIdempotencyStore) Commit(
	ctx context.Context,
	lock idempotency.IdempotencyLock,
	resultJSON string,
) error {
	return f.commit(ctx, lock, resultJSON)
}

func (f *fakeIdempotencyStore) Release(ctx context.Context, lock idempotency.IdempotencyLock) error {
	return f.release(ctx, lock)
}

func TestIdempotencyAccountRepository_AdjustBalance(t *testing.T) {
	delta, err := valueobject.NewDecimal("50")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}
	lock := idempotency.IdempotencyLock{Key: "key-1", Token: "token-1"}

	t.Run("acquired success", func(t *testing.T) {
		innerCalls := 0
		commitCalls := 0
		releaseCalls := 0

		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return lock, idempotency.IdempotencyStatusAcquired, "", nil
			},
			commit: func(ctx context.Context, lock idempotency.IdempotencyLock, resultJSON string) error {
				commitCalls++
				return nil
			},
			release: func(ctx context.Context, lock idempotency.IdempotencyLock) error {
				releaseCalls++
				return nil
			},
		}

		repository := NewIdempotencyAccountRepository(inner, store)

		if err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if innerCalls != 1 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 1)
		}

		if commitCalls != 1 {
			t.Errorf("got %d commit calls, want %d", commitCalls, 1)
		}

		if releaseCalls != 0 {
			t.Errorf("got %d release calls, want %d", releaseCalls, 0)
		}
	})

	t.Run("acquired inner error releases", func(t *testing.T) {
		boom := errors.New("boom")
		releaseCalls := 0
		commitCalls := 0

		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				return boom
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return lock, idempotency.IdempotencyStatusAcquired, "", nil
			},
			commit: func(ctx context.Context, lock idempotency.IdempotencyLock, resultJSON string) error {
				commitCalls++
				return nil
			},
			release: func(ctx context.Context, lock idempotency.IdempotencyLock) error {
				releaseCalls++
				return nil
			},
		}

		repository := NewIdempotencyAccountRepository(inner, store)

		err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if releaseCalls != 1 {
			t.Errorf("got %d release calls, want %d", releaseCalls, 1)
		}

		if commitCalls != 0 {
			t.Errorf("got %d commit calls, want %d", commitCalls, 0)
		}
	})

	t.Run("locked", func(t *testing.T) {
		innerCalls := 0

		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return idempotency.IdempotencyLock{}, idempotency.IdempotencyStatusLocked, "", nil
			},
		}

		repository := NewIdempotencyAccountRepository(inner, store)

		err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

		if !errors.Is(err, idempotency.ErrResourceLocked) {
			t.Errorf("got %v, want %v", err, idempotency.ErrResourceLocked)
		}

		if innerCalls != 0 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 0)
		}
	})

	t.Run("completed", func(t *testing.T) {
		innerCalls := 0

		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return idempotency.IdempotencyLock{}, idempotency.IdempotencyStatusCompleted, "", nil
			},
		}

		repository := NewIdempotencyAccountRepository(inner, store)

		if err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if innerCalls != 0 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 0)
		}
	})

	t.Run("acquire error", func(t *testing.T) {
		boom := errors.New("boom")
		innerCalls := 0

		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return idempotency.IdempotencyLock{}, "", "", boom
			},
		}

		repository := NewIdempotencyAccountRepository(inner, store)

		err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if innerCalls != 0 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 0)
		}
	})
}
