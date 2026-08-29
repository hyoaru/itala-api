package transaction

import (
	"context"
	"errors"
	"math"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
)

type RetryTransactionRepository struct {
	inner       port.TransactionRepository
	maxAttempts int8
	delay       time.Duration
	maxDelay    time.Duration
}

func NewRetryTransactionRepository(inner port.TransactionRepository, maxAttempts int8, delay time.Duration, maxDelay time.Duration) *RetryTransactionRepository {
	return &RetryTransactionRepository{inner: inner, maxAttempts: maxAttempts, delay: delay, maxDelay: maxDelay}
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

func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	if errors.Is(err, port.ErrTransactionExists) {
		return false
	}

	if errors.Is(err, port.ErrTransactionNotFound) {
		return false
	}

	return true
}

func (r *RetryTransactionRepository) backoff(attempt int8) time.Duration {
	delay := r.delay * time.Duration(math.Pow(2, float64(attempt)))

	if delay > r.maxDelay {
		return r.maxDelay
	}

	return delay
}

func (r *RetryTransactionRepository) Create(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Create(ctx, userID, transaction, idempotencyKey); err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return err
		}
	}
	return err
}

func (r *RetryTransactionRepository) Find(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error) {
	var (
		transactionPage port.TransactionPage
		err             error
	)

	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		transactionPage, err = r.inner.Find(ctx, userID, query)
		if err == nil {
			return transactionPage, nil
		}

		if !isRetryable(err) {
			return port.TransactionPage{}, err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return port.TransactionPage{}, err
		}
	}

	return port.TransactionPage{}, err
}

func (r *RetryTransactionRepository) FindOne(ctx context.Context, userID string, id string) (entity.Transaction, error) {
	var (
		transaction entity.Transaction
		err         error
	)

	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		transaction, err = r.inner.FindOne(ctx, userID, id)
		if err == nil {
			return transaction, nil
		}

		if !isRetryable(err) {
			return entity.Transaction{}, err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return entity.Transaction{}, err
		}
	}

	return transaction, err
}

func (r *RetryTransactionRepository) Update(ctx context.Context, userID string, transaction entity.Transaction) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Update(ctx, userID, transaction); err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return err
		}
	}

	return err
}

func (r *RetryTransactionRepository) Delete(ctx context.Context, userID string, id string) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Delete(ctx, userID, id); err == nil {
			return nil
		}

		if !isRetryable(err) {
			return err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return err
		}
	}
	return err
}
