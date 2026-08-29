package idempotency

import (
	"context"
	"math"
	"time"
)

type RetryIdempotencyStore struct {
	inner       IdempotencyStore
	maxAttempts int8
	delay       time.Duration
	maxDelay    time.Duration
}

func NewRetryIdempotencyStore(inner IdempotencyStore, maxAttempts int8, delay time.Duration, maxDelay time.Duration) IdempotencyStore {
	return &RetryIdempotencyStore{inner: inner, maxAttempts: maxAttempts, delay: delay, maxDelay: maxDelay}
}

func sleep(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *RetryIdempotencyStore) backoff(attempt int8) time.Duration {
	delay := s.delay * time.Duration(math.Pow(2, float64(attempt)))

	if delay > s.maxDelay {
		return s.maxDelay
	}

	return delay
}

func (s *RetryIdempotencyStore) Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
	var (
		lock   IdempotencyLock
		status IdempotencyStatus
		result ResultJSON
		err    error
	)

	for attempt := int8(0); attempt < s.maxAttempts; attempt++ {
		lock, status, result, err = s.inner.Acquire(ctx, key, ttl)
		if err == nil {
			return lock, status, result, nil
		}

		if attempt == s.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, s.backoff(attempt)); err != nil {
			return IdempotencyLock{}, "", "", err
		}
	}

	return IdempotencyLock{}, "", "", err
}

func (s *RetryIdempotencyStore) Commit(ctx context.Context, lock IdempotencyLock, result string) error {
	var err error
	for attempt := int8(0); attempt < s.maxAttempts; attempt++ {
		if err = s.inner.Commit(ctx, lock, result); err == nil {
			return nil
		}

		if attempt == s.maxAttempts-1 {
			break
		}

		if err := sleep(ctx, s.backoff(attempt)); err != nil {
			return err
		}
	}

	return err
}

func (s *RetryIdempotencyStore) Release(ctx context.Context, lock IdempotencyLock) error {
	var err error
	for attempt := int8(0); attempt < s.maxAttempts; attempt++ {
		if err = s.inner.Release(ctx, lock); err == nil {
			return nil
		}

		if attempt == s.maxAttempts-1 {
			break
		}

		if err := sleep(ctx, s.backoff(attempt)); err != nil {
			return err
		}
	}

	return err
}
