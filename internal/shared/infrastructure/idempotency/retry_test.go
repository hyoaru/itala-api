package idempotency

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubIdempotencyStore struct {
	IdempotencyStore
	acquire func(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error)
	commit  func(ctx context.Context, lock IdempotencyLock, result string) error
	release func(ctx context.Context, lock IdempotencyLock) error
}

func (f *stubIdempotencyStore) Acquire(
	ctx context.Context,
	key string,
	expiresAt uint16,
) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
	return f.acquire(ctx, key, expiresAt)
}

func (f *stubIdempotencyStore) Commit(ctx context.Context, lock IdempotencyLock, result string) error {
	return f.commit(ctx, lock, result)
}

func (f *stubIdempotencyStore) Release(ctx context.Context, lock IdempotencyLock) error {
	return f.release(ctx, lock)
}

func newRetryStore(inner IdempotencyStore) IdempotencyStore {
	return NewRetryIdempotencyStore(inner, 3, time.Millisecond, 2*time.Millisecond)
}

func TestIsRetryable(t *testing.T) {
	t.Run("context canceled", func(t *testing.T) {
		if isRetryable(context.Canceled) {
			t.Error("got true, want false")
		}
	})

	t.Run("context deadline exceeded", func(t *testing.T) {
		if isRetryable(context.DeadlineExceeded) {
			t.Error("got true, want false")
		}
	})

	t.Run("invalid lock token is retryable", func(t *testing.T) {
		if !isRetryable(ErrInvalidLockToken) {
			t.Error("got false, want true")
		}
	})

	t.Run("item not found is retryable", func(t *testing.T) {
		if !isRetryable(ErrItemNotFound) {
			t.Error("got false, want true")
		}
	})

	t.Run("generic error", func(t *testing.T) {
		if !isRetryable(errors.New("boom")) {
			t.Error("got false, want true")
		}
	})
}

func TestRetryIdempotencyStore_backoff(t *testing.T) {
	repository := NewRetryIdempotencyStore(nil, 5, 100*time.Millisecond, 2*time.Second).(*RetryIdempotencyStore)

	t.Run("first attempt", func(t *testing.T) {
		if got := repository.backoff(0); got != 100*time.Millisecond {
			t.Errorf("got %v, want %v", got, 100*time.Millisecond)
		}
	})

	t.Run("doubles per attempt", func(t *testing.T) {
		if got := repository.backoff(2); got != 400*time.Millisecond {
			t.Errorf("got %v, want %v", got, 400*time.Millisecond)
		}
	})

	t.Run("caps at max delay", func(t *testing.T) {
		if got := repository.backoff(5); got != 2*time.Second {
			t.Errorf("got %v, want %v", got, 2*time.Second)
		}
	})
}

func TestRetryIdempotencyStore_Acquire(t *testing.T) {
	lock := IdempotencyLock{Key: "key-1", Token: "token-1"}

	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
				calls++
				return lock, IdempotencyStatusAcquired, "result-1", nil
			},
		}

		store := newRetryStore(inner)

		gotLock, gotStatus, gotResult, err := store.Acquire(context.Background(), "key-1", 900)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if gotLock != lock || gotStatus != IdempotencyStatusAcquired || gotResult != "result-1" {
			t.Errorf("got %v/%v/%v, want %v/%v/%v", gotLock, gotStatus, gotResult, lock, IdempotencyStatusAcquired, "result-1")
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
				calls++
				if calls == 1 {
					return IdempotencyLock{}, "", "", errors.New("boom")
				}
				return lock, IdempotencyStatusAcquired, "", nil
			},
		}

		store := newRetryStore(inner)

		if _, _, _, err := store.Acquire(context.Background(), "key-1", 900); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
				calls++
				return lock, IdempotencyStatusAcquired, "result-1", context.Canceled
			},
		}

		store := newRetryStore(inner)

		gotLock, gotStatus, gotResult, err := store.Acquire(context.Background(), "key-1", 900)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v, want %v", err, context.Canceled)
		}

		if gotLock != (IdempotencyLock{}) || gotStatus != "" || gotResult != "" {
			t.Errorf("got %v/%v/%v, want zero values", gotLock, gotStatus, gotResult)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("exhausted attempts returns last error", func(t *testing.T) {
		boom := errors.New("boom")
		calls := 0
		inner := &stubIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
				calls++
				return IdempotencyLock{}, "", "", boom
			},
		}

		store := newRetryStore(inner)

		_, _, _, err := store.Acquire(context.Background(), "key-1", 900)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if calls != 3 {
			t.Errorf("got %d calls, want %d", calls, 3)
		}
	})

	t.Run("context canceled during sleep", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			acquire: func(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
				calls++
				return IdempotencyLock{}, "", "", errors.New("boom")
			},
		}

		store := newRetryStore(inner)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, _, _, err := store.Acquire(ctx, "key-1", 900)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v, want %v", err, context.Canceled)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryIdempotencyStore_Commit(t *testing.T) {
	lock := IdempotencyLock{Key: "key-1", Token: "token-1"}

	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			commit: func(ctx context.Context, lock IdempotencyLock, result string) error {
				calls++
				return nil
			},
		}

		store := newRetryStore(inner)

		if err := store.Commit(context.Background(), lock, "result-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			commit: func(ctx context.Context, lock IdempotencyLock, result string) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		store := newRetryStore(inner)

		if err := store.Commit(context.Background(), lock, "result-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			commit: func(ctx context.Context, lock IdempotencyLock, result string) error {
				calls++
				return context.Canceled
			},
		}

		store := newRetryStore(inner)

		err := store.Commit(context.Background(), lock, "result-1")

		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v, want %v", err, context.Canceled)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryIdempotencyStore_Release(t *testing.T) {
	lock := IdempotencyLock{Key: "key-1", Token: "token-1"}

	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			release: func(ctx context.Context, lock IdempotencyLock) error {
				calls++
				return nil
			},
		}

		store := newRetryStore(inner)

		if err := store.Release(context.Background(), lock); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			release: func(ctx context.Context, lock IdempotencyLock) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		store := newRetryStore(inner)

		if err := store.Release(context.Background(), lock); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubIdempotencyStore{
			release: func(ctx context.Context, lock IdempotencyLock) error {
				calls++
				return context.Canceled
			},
		}

		store := newRetryStore(inner)

		err := store.Release(context.Background(), lock)

		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v, want %v", err, context.Canceled)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}
