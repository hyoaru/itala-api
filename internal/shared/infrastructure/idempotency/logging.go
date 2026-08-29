package idempotency

import (
	"context"

	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingIdempotencyStore struct {
	inner IdempotencyStore
}

func NewLoggingIdempotencyStore(inner IdempotencyStore) *LoggingIdempotencyStore {
	return &LoggingIdempotencyStore{inner: inner}
}

func (l *LoggingIdempotencyStore) Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
	logger.Debug("Acquiring idempotency lock", "key", key, "ttl", ttl)
	lock, status, result, err := l.inner.Acquire(ctx, key, ttl)
	if err != nil {
		logger.Warn("Failed to acquire idempotency lock", "error", err)
		return lock, status, result, err
	}
	logger.Debug("Idempotency lock acquired", "key", key, "ttl", ttl)
	return lock, status, result, nil
}

func (l *LoggingIdempotencyStore) Commit(ctx context.Context, lock IdempotencyLock, result string) error {
	logger.Debug("Committing idempotency lock", "key", lock.Key, "token", lock.Token)
	if err := l.inner.Commit(ctx, lock, result); err != nil {
		logger.Warn("Failed to commit idempotency lock", "error", err)
		return err
	}
	logger.Debug("Idempotency lock committed", "key", lock.Key, "token", lock.Token)
	return nil
}

func (l *LoggingIdempotencyStore) Release(ctx context.Context, lock IdempotencyLock) error {
	logger.Debug("Releasing idempotency lock", "key", lock.Key, "token", lock.Token)
	if err := l.inner.Release(ctx, lock); err != nil {
		logger.Warn("Failed to release idempotency lock", "error", err)
		return err
	}
	logger.Debug("Idempotency lock released", "key", lock.Key, "token", lock.Token)
	return nil
}
