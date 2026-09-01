package idempotency

import (
	"context"
)

type (
	IdempotencyStatus string
	ResultJSON        string
)

const (
	IdempotencyStatusAcquired  IdempotencyStatus = "ACQUIRED"
	IdempotencyStatusLocked    IdempotencyStatus = "LOCKED"
	IdempotencyStatusCompleted IdempotencyStatus = "COMPLETED"
)

type IdempotencyLock struct {
	Key   string
	Token string
}

type IdempotencyStore interface {
	Acquire(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error)
	Commit(ctx context.Context, Lock IdempotencyLock, resultJSON string) error
	Release(ctx context.Context, Lock IdempotencyLock) error
}
