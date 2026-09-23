package transaction

import (
	"context"
	"errors"
	"testing"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
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

func TestIdempotencyTransactionRepository_Create(t *testing.T) {
	lock := idempotency.IdempotencyLock{Key: "idem-1", Token: "token-1"}

	t.Run("acquired success", func(t *testing.T) {
		innerCalls := 0
		commitCalls := 0
		releaseCalls := 0

		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
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

		repository := NewIdempotencyTransactionRepository(inner, store)

		if err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1"); err != nil {
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

	t.Run("acquired transaction exists does not release", func(t *testing.T) {
		commitCalls := 0
		releaseCalls := 0

		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				return port.ErrTransactionExists
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

		repository := NewIdempotencyTransactionRepository(inner, store)

		err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(err, port.ErrTransactionExists) {
			t.Errorf("got %v, want %v", err, port.ErrTransactionExists)
		}

		if releaseCalls != 0 {
			t.Errorf("got %d release calls, want %d", releaseCalls, 0)
		}

		if commitCalls != 0 {
			t.Errorf("got %d commit calls, want %d", commitCalls, 0)
		}
	})

	t.Run("acquired other error releases", func(t *testing.T) {
		boom := errors.New("boom")
		commitCalls := 0
		releaseCalls := 0

		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
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

		repository := NewIdempotencyTransactionRepository(inner, store)

		err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

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

		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return idempotency.IdempotencyLock{}, idempotency.IdempotencyStatusLocked, "", nil
			},
		}

		repository := NewIdempotencyTransactionRepository(inner, store)

		err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(err, idempotency.ErrResourceLocked) {
			t.Errorf("got %v, want %v", err, idempotency.ErrResourceLocked)
		}

		if innerCalls != 0 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 0)
		}
	})

	t.Run("completed", func(t *testing.T) {
		innerCalls := 0

		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return idempotency.IdempotencyLock{}, idempotency.IdempotencyStatusCompleted, "", nil
			},
		}

		repository := NewIdempotencyTransactionRepository(inner, store)

		if err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if innerCalls != 0 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 0)
		}
	})

	t.Run("acquire error", func(t *testing.T) {
		boom := errors.New("boom")
		innerCalls := 0

		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				innerCalls++
				return nil
			},
		}
		store := &fakeIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (idempotency.IdempotencyLock, idempotency.IdempotencyStatus, idempotency.ResultJSON, error) {
				return idempotency.IdempotencyLock{}, "", "", boom
			},
		}

		repository := NewIdempotencyTransactionRepository(inner, store)

		err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if innerCalls != 0 {
			t.Errorf("got %d inner calls, want %d", innerCalls, 0)
		}
	})
}
