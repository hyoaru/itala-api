package account

import (
	"context"
	"errors"
	"math"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type RetryAccountRepository struct {
	inner       port.AccountRepository
	maxAttempts int8
	delay       time.Duration
	maxDelay    time.Duration
}

func NewRetryAccountRepository(inner port.AccountRepository, maxAttempts int8, delay time.Duration, maxDelay time.Duration) *RetryAccountRepository {
	return &RetryAccountRepository{inner: inner, maxAttempts: maxAttempts, delay: delay, maxDelay: maxDelay}
}

func isRetryable(err error) bool {
	if errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	if errors.Is(err, port.ErrAccountExists) ||
		errors.Is(err, port.ErrAccountNotFound) ||
		errors.Is(err, port.ErrAccountArchived) ||
		errors.Is(err, port.ErrConcurrentModification) {
		return false
	}

	return true
}

func (r *RetryAccountRepository) backoff(attempt int8) time.Duration {
	delay := r.delay * time.Duration(math.Pow(2, float64(attempt)))

	if delay > r.maxDelay {
		return r.maxDelay
	}

	return delay
}

func (r *RetryAccountRepository) AdjustBalance(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.AdjustBalance(ctx, userID, accountID, idempotencyKey, delta); err == nil {
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

func (r *RetryAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Create(ctx, userID, account); err == nil {
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

func (r *RetryAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	var (
		accountPage port.AccountPage
		err         error
	)

	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		accountPage, err = r.inner.Find(ctx, userID, query)
		if err == nil {
			return accountPage, nil
		}

		if !isRetryable(err) {
			return port.AccountPage{}, err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return port.AccountPage{}, err
		}
	}

	return port.AccountPage{}, err
}

func (r *RetryAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	var (
		account entity.Account
		err     error
	)

	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		account, err = r.inner.FindOne(ctx, userID, id)
		if err == nil {
			return account, nil
		}

		if !isRetryable(err) {
			return entity.Account{}, err
		}

		if attempt == r.maxAttempts-1 {
			break
		}

		if err = sleep(ctx, r.backoff(attempt)); err != nil {
			return entity.Account{}, err
		}
	}

	return entity.Account{}, err
}

func (r *RetryAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Update(ctx, userID, account); err == nil {
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

func (r *RetryAccountRepository) Archive(ctx context.Context, userID string, id string) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Archive(ctx, userID, id); err == nil {
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

func (r *RetryAccountRepository) Restore(ctx context.Context, userID string, id string) error {
	var err error
	for attempt := int8(0); attempt < r.maxAttempts; attempt++ {
		if err = r.inner.Restore(ctx, userID, id); err == nil {
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
