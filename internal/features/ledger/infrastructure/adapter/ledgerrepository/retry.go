package ledgerrepository

import (
	"context"
	"errors"
	"math"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
)

type RetryLedgerRepository struct {
	inner       port.LedgerRepository
	maxAttempts int8
	delay       time.Duration
	maxDelay    time.Duration
}

func NewRetryLedgerRepository(inner port.LedgerRepository) *RetryLedgerRepository {
	return &RetryLedgerRepository{inner: inner, maxAttempts: 5, delay: 100 * time.Millisecond, maxDelay: 2 * time.Second}
}

func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return !errors.Is(err, port.ErrConditionFailed) &&
		!errors.Is(err, port.ErrIdempotencyMismatch) &&
		!errors.Is(err, port.ErrTransactionNotFound)
}

func (r *RetryLedgerRepository) retry(ctx context.Context, operation func() error) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = operation(); err == nil {
			return nil
		}
		if !isRetryable(err) || attempt == r.maxAttempts-1 {
			return err
		}

		delay := r.delay * time.Duration(math.Pow(2, float64(attempt)))
		if delay > r.maxDelay {
			delay = r.maxDelay
		}

		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}

func (r *RetryLedgerRepository) RecordTransaction(ctx context.Context, userID string, input port.RecordTransactionInput) error {
	return r.retry(ctx, func() error { return r.inner.RecordTransaction(ctx, userID, input) })
}

func (r *RetryLedgerRepository) FindIdempotency(ctx context.Context, userID string, keyHash string) (port.IdempotencyRecord, bool, error) {
	var record port.IdempotencyRecord
	var found bool
	err := r.retry(ctx, func() error {
		var err error
		record, found, err = r.inner.FindIdempotency(ctx, userID, keyHash)
		return err
	})
	return record, found, err
}

func (r *RetryLedgerRepository) FindTransactionDetail(ctx context.Context, userID string, occurredAt string, id string) (entity.Transaction, []entity.Posting, error) {
	var transaction entity.Transaction
	var postings []entity.Posting
	err := r.retry(ctx, func() error {
		var err error
		transaction, postings, err = r.inner.FindTransactionDetail(ctx, userID, occurredAt, id)
		return err
	})
	return transaction, postings, err
}
