package accountrepository

import (
	"context"
	"errors"
	"math"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type RetryAccountRepository struct {
	inner       port.AccountRepository
	maxAttempts int8
	delay       time.Duration
	maxDelay    time.Duration
}

func NewRetryAccountRepository(inner port.AccountRepository) *RetryAccountRepository {
	return &RetryAccountRepository{inner: inner, maxAttempts: 5, delay: 100 * time.Millisecond, maxDelay: 2 * time.Second}
}

func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	return !errors.Is(err, port.ErrAccountExists) &&
		!errors.Is(err, port.ErrAccountNotFound) &&
		!errors.Is(err, port.ErrConcurrentModification)
}

func (r *RetryAccountRepository) retry(ctx context.Context, operation func() error) error {
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

func (r *RetryAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	return r.retry(ctx, func() error { return r.inner.Create(ctx, userID, account) })
}

func (r *RetryAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	var page port.AccountPage
	err := r.retry(ctx, func() error {
		var err error
		page, err = r.inner.Find(ctx, userID, query)
		return err
	})
	return page, err
}

func (r *RetryAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	var account entity.Account
	err := r.retry(ctx, func() error {
		var err error
		account, err = r.inner.FindOne(ctx, userID, id)
		return err
	})
	return account, err
}

func (r *RetryAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	return r.retry(ctx, func() error { return r.inner.Update(ctx, userID, account) })
}

func (r *RetryAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	return r.retry(ctx, func() error { return r.inner.Delete(ctx, userID, id) })
}
