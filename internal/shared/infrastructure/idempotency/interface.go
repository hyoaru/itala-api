package idempotency

import "context"

type IdempotencyStatus string

const (
	IdempotencyStatusAcquired  IdempotencyStatus = "ACQUIRED"
	IdempotencyStatusLocked    IdempotencyStatus = "LOCKED"
	IdempotencyStatusCompleted IdempotencyStatus = "COMPLETED"
)

type IdempotencyLock struct {
	Status IdempotencyStatus
	Result any
}

type IdempotencyStore interface {
	Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, error)
	Commit(ctx context.Context, key string, result any) error
	Release(ctx context.Context, key string) error
}
