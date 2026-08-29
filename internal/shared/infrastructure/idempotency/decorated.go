package idempotency

import (
	"context"
	"time"
)

type DecoratedIdempotencyStore struct {
	inner IdempotencyStore
}

func NewDecoratedIdempotencyStore(inner IdempotencyStore) *DecoratedIdempotencyStore {
	logging := NewLoggingIdempotencyStore(inner)
	retry := NewRetryIdempotencyStore(logging, 5, 1*time.Second)
	return &DecoratedIdempotencyStore{inner: retry}
}

func (d *DecoratedIdempotencyStore) Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
	return d.inner.Acquire(ctx, key, ttl)
}

func (d *DecoratedIdempotencyStore) Commit(ctx context.Context, lock IdempotencyLock, result string) error {
	return d.inner.Commit(ctx, lock, result)
}

func (d *DecoratedIdempotencyStore) Release(ctx context.Context, lock IdempotencyLock) error {
	return d.inner.Release(ctx, lock)
}
