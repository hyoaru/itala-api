package account

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type stubAccountRepository struct {
	port.AccountRepository
	create        func(ctx context.Context, userID string, account entity.Account) error
	find          func(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error)
	findOne       func(ctx context.Context, userID string, id string) (entity.Account, error)
	update        func(ctx context.Context, userID string, account entity.Account) error
	delete        func(ctx context.Context, userID string, id string) error
	adjustBalance func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error
}

func (f *stubAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	return f.create(ctx, userID, account)
}

func (f *stubAccountRepository) Find(
	ctx context.Context,
	userID string,
	query port.AccountQuery,
) (port.AccountPage, error) {
	return f.find(ctx, userID, query)
}

func (f *stubAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	return f.findOne(ctx, userID, id)
}

func (f *stubAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	return f.update(ctx, userID, account)
}

func (f *stubAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	return f.delete(ctx, userID, id)
}

func (f *stubAccountRepository) AdjustBalance(
	ctx context.Context,
	userID string,
	accountID string,
	idempotencyKey string,
	delta valueobject.Decimal,
) error {
	return f.adjustBalance(ctx, userID, accountID, idempotencyKey, delta)
}

func newRetryRepository(inner port.AccountRepository) *RetryAccountRepository {
	return NewRetryAccountRepository(inner, 3, time.Millisecond, 2*time.Millisecond)
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

	t.Run("account exists", func(t *testing.T) {
		if isRetryable(port.ErrAccountExists) {
			t.Error("got true, want false")
		}
	})

	t.Run("account not found", func(t *testing.T) {
		if isRetryable(port.ErrAccountNotFound) {
			t.Error("got true, want false")
		}
	})

	t.Run("account deleted", func(t *testing.T) {
		if isRetryable(port.ErrAccountDeleted) {
			t.Error("got true, want false")
		}
	})

	t.Run("concurrent modification", func(t *testing.T) {
		if isRetryable(port.ErrConcurrentModification) {
			t.Error("got true, want false")
		}
	})

	t.Run("generic error", func(t *testing.T) {
		if !isRetryable(errors.New("boom")) {
			t.Error("got false, want true")
		}
	})
}

func TestRetryAccountRepository_backoff(t *testing.T) {
	repository := NewRetryAccountRepository(nil, 5, 100*time.Millisecond, 2*time.Second)

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

func TestRetryAccountRepository_Create(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Create(context.Background(), "user-1", entity.Account{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Create(context.Background(), "user-1", entity.Account{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				return port.ErrAccountExists
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Create(context.Background(), "user-1", entity.Account{})

		if !errors.Is(err, port.ErrAccountExists) {
			t.Errorf("got %v, want %v", err, port.ErrAccountExists)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("exhausted attempts returns last error", func(t *testing.T) {
		boom := errors.New("boom")
		calls := 0
		inner := &stubAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				return boom
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Create(context.Background(), "user-1", entity.Account{})

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if calls != 3 {
			t.Errorf("got %d calls, want %d", calls, 3)
		}
	})

	t.Run("context canceled during sleep", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				return errors.New("boom")
			},
		}

		repository := newRetryRepository(inner)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := repository.Create(ctx, "user-1", entity.Account{})

		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v, want %v", err, context.Canceled)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryAccountRepository_Find(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			find: func(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
				calls++
				return port.AccountPage{Accounts: []entity.Account{{ID: "account-1"}}}, nil
			},
		}

		repository := newRetryRepository(inner)

		got, err := repository.Find(context.Background(), "user-1", port.AccountQuery{})
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if len(got.Accounts) != 1 {
			t.Errorf("got %d accounts, want %d", len(got.Accounts), 1)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			find: func(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
				calls++
				if calls == 1 {
					return port.AccountPage{}, errors.New("boom")
				}
				return port.AccountPage{Accounts: []entity.Account{{ID: "account-1"}}}, nil
			},
		}

		repository := newRetryRepository(inner)

		if _, err := repository.Find(context.Background(), "user-1", port.AccountQuery{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			find: func(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
				calls++
				return port.AccountPage{}, port.ErrAccountNotFound
			},
		}

		repository := newRetryRepository(inner)

		_, err := repository.Find(context.Background(), "user-1", port.AccountQuery{})

		if !errors.Is(err, port.ErrAccountNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrAccountNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryAccountRepository_FindOne(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				calls++
				return entity.Account{ID: "account-1"}, nil
			},
		}

		repository := newRetryRepository(inner)

		got, err := repository.FindOne(context.Background(), "user-1", "account-1")
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "account-1" {
			t.Errorf("got %v, want %v", got.ID, "account-1")
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				calls++
				if calls == 1 {
					return entity.Account{}, errors.New("boom")
				}
				return entity.Account{ID: "account-1"}, nil
			},
		}

		repository := newRetryRepository(inner)

		if _, err := repository.FindOne(context.Background(), "user-1", "account-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				calls++
				return entity.Account{}, port.ErrAccountNotFound
			},
		}

		repository := newRetryRepository(inner)

		_, err := repository.FindOne(context.Background(), "user-1", "account-1")

		if !errors.Is(err, port.ErrAccountNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrAccountNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryAccountRepository_Update(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			update: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Update(context.Background(), "user-1", entity.Account{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			update: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Update(context.Background(), "user-1", entity.Account{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			update: func(ctx context.Context, userID string, account entity.Account) error {
				calls++
				return port.ErrConcurrentModification
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Update(context.Background(), "user-1", entity.Account{})

		if !errors.Is(err, port.ErrConcurrentModification) {
			t.Errorf("got %v, want %v", err, port.ErrConcurrentModification)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryAccountRepository_Delete(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Delete(context.Background(), "user-1", "account-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Delete(context.Background(), "user-1", "account-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				calls++
				return port.ErrAccountNotFound
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Delete(context.Background(), "user-1", "account-1")

		if !errors.Is(err, port.ErrAccountNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrAccountNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryAccountRepository_AdjustBalance(t *testing.T) {
	delta, err := valueobject.NewDecimal("50")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}

	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubAccountRepository{
			adjustBalance: func(
				ctx context.Context,
				userID string,
				accountID string,
				idempotencyKey string,
				delta valueobject.Decimal,
			) error {
				calls++
				return port.ErrAccountNotFound
			},
		}

		repository := newRetryRepository(inner)

		err := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

		if !errors.Is(err, port.ErrAccountNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrAccountNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}
