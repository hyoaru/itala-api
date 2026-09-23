package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
)

type stubTransactionRepository struct {
	port.TransactionRepository
	create  func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error
	find    func(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error)
	findOne func(ctx context.Context, userID string, id string) (entity.Transaction, error)
	update  func(ctx context.Context, userID string, transaction entity.Transaction) error
	delete  func(ctx context.Context, userID string, id string) error
}

func (f *stubTransactionRepository) Create(
	ctx context.Context,
	userID string,
	transaction entity.Transaction,
	idempotencyKey string,
) error {
	return f.create(ctx, userID, transaction, idempotencyKey)
}

func (f *stubTransactionRepository) Find(
	ctx context.Context,
	userID string,
	query port.TransactionQuery,
) (port.TransactionPage, error) {
	return f.find(ctx, userID, query)
}

func (f *stubTransactionRepository) FindOne(ctx context.Context, userID string, id string) (entity.Transaction, error) {
	return f.findOne(ctx, userID, id)
}

func (f *stubTransactionRepository) Update(ctx context.Context, userID string, transaction entity.Transaction) error {
	return f.update(ctx, userID, transaction)
}

func (f *stubTransactionRepository) Delete(ctx context.Context, userID string, id string) error {
	return f.delete(ctx, userID, id)
}

func newRetryRepository(inner port.TransactionRepository) *RetryTransactionRepository {
	return NewRetryTransactionRepository(inner, 3, time.Millisecond, 2*time.Millisecond)
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

	t.Run("transaction exists", func(t *testing.T) {
		if isRetryable(port.ErrTransactionExists) {
			t.Error("got true, want false")
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		if isRetryable(port.ErrTransactionNotFound) {
			t.Error("got true, want false")
		}
	})

	t.Run("concurrent modification is retryable", func(t *testing.T) {
		if !isRetryable(port.ErrConcurrentModification) {
			t.Error("got false, want true")
		}
	})

	t.Run("generic error", func(t *testing.T) {
		if !isRetryable(errors.New("boom")) {
			t.Error("got false, want true")
		}
	})
}

func TestRetryTransactionRepository_backoff(t *testing.T) {
	repository := NewRetryTransactionRepository(nil, 5, 100*time.Millisecond, 2*time.Second)

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

func TestRetryTransactionRepository_Create(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				calls++
				return port.ErrTransactionExists
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(err, port.ErrTransactionExists) {
			t.Errorf("got %v, want %v", err, port.ErrTransactionExists)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("exhausted attempts returns last error", func(t *testing.T) {
		boom := errors.New("boom")
		calls := 0
		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				calls++
				return boom
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if calls != 3 {
			t.Errorf("got %d calls, want %d", calls, 3)
		}
	})

	t.Run("context canceled during sleep", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				calls++
				return errors.New("boom")
			},
		}

		repository := newRetryRepository(inner)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := repository.Create(ctx, "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(err, context.Canceled) {
			t.Errorf("got %v, want %v", err, context.Canceled)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryTransactionRepository_Find(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			find: func(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error) {
				calls++
				return port.TransactionPage{Transactions: []entity.Transaction{{ID: "transaction-1"}}}, nil
			},
		}

		repository := newRetryRepository(inner)

		got, err := repository.Find(context.Background(), "user-1", port.TransactionQuery{})
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if len(got.Transactions) != 1 {
			t.Errorf("got %d transactions, want %d", len(got.Transactions), 1)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			find: func(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error) {
				calls++
				if calls == 1 {
					return port.TransactionPage{}, errors.New("boom")
				}
				return port.TransactionPage{Transactions: []entity.Transaction{{ID: "transaction-1"}}}, nil
			},
		}

		repository := newRetryRepository(inner)

		if _, err := repository.Find(context.Background(), "user-1", port.TransactionQuery{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			find: func(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error) {
				calls++
				return port.TransactionPage{}, port.ErrTransactionNotFound
			},
		}

		repository := newRetryRepository(inner)

		_, err := repository.Find(context.Background(), "user-1", port.TransactionQuery{})

		if !errors.Is(err, port.ErrTransactionNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrTransactionNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryTransactionRepository_FindOne(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				calls++
				return entity.Transaction{ID: "transaction-1"}, nil
			},
		}

		repository := newRetryRepository(inner)

		got, err := repository.FindOne(context.Background(), "user-1", "transaction-1")
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "transaction-1" {
			t.Errorf("got %v, want %v", got.ID, "transaction-1")
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				calls++
				if calls == 1 {
					return entity.Transaction{}, errors.New("boom")
				}
				return entity.Transaction{ID: "transaction-1"}, nil
			},
		}

		repository := newRetryRepository(inner)

		if _, err := repository.FindOne(context.Background(), "user-1", "transaction-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				calls++
				return entity.Transaction{}, port.ErrTransactionNotFound
			},
		}

		repository := newRetryRepository(inner)

		_, err := repository.FindOne(context.Background(), "user-1", "transaction-1")

		if !errors.Is(err, port.ErrTransactionNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrTransactionNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryTransactionRepository_Update(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			update: func(ctx context.Context, userID string, transaction entity.Transaction) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Update(context.Background(), "user-1", entity.Transaction{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			update: func(ctx context.Context, userID string, transaction entity.Transaction) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Update(context.Background(), "user-1", entity.Transaction{}); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			update: func(ctx context.Context, userID string, transaction entity.Transaction) error {
				calls++
				return port.ErrTransactionNotFound
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Update(context.Background(), "user-1", entity.Transaction{})

		if !errors.Is(err, port.ErrTransactionNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrTransactionNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}

func TestRetryTransactionRepository_Delete(t *testing.T) {
	t.Run("success first attempt", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				calls++
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Delete(context.Background(), "user-1", "transaction-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})

	t.Run("retryable then success", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				calls++
				if calls == 1 {
					return errors.New("boom")
				}
				return nil
			},
		}

		repository := newRetryRepository(inner)

		if err := repository.Delete(context.Background(), "user-1", "transaction-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if calls != 2 {
			t.Errorf("got %d calls, want %d", calls, 2)
		}
	})

	t.Run("non-retryable returns immediately", func(t *testing.T) {
		calls := 0
		inner := &stubTransactionRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				calls++
				return port.ErrTransactionNotFound
			},
		}

		repository := newRetryRepository(inner)

		err := repository.Delete(context.Background(), "user-1", "transaction-1")

		if !errors.Is(err, port.ErrTransactionNotFound) {
			t.Errorf("got %v, want %v", err, port.ErrTransactionNotFound)
		}

		if calls != 1 {
			t.Errorf("got %d calls, want %d", calls, 1)
		}
	})
}
