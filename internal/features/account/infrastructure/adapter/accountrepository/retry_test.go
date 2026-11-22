package accountrepository

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type stubAccountRepository struct {
	port.AccountRepository
	accounts     map[string]entity.Account
	createCalls  int
	createErrors []error
}

func newStubAccountRepository() *stubAccountRepository {
	return &stubAccountRepository{accounts: map[string]entity.Account{}}
}

func (f *stubAccountRepository) Create(_ context.Context, _ string, account entity.Account) error {
	f.createCalls++
	if len(f.createErrors) > 0 {
		err := f.createErrors[0]
		f.createErrors = f.createErrors[1:]
		if err != nil {
			return err
		}
	}
	f.accounts[account.ID] = account
	return nil
}

func (f *stubAccountRepository) FindOne(_ context.Context, _ string, id string) (entity.Account, error) {
	account, ok := f.accounts[id]
	if !ok {
		return entity.Account{}, port.ErrAccountNotFound
	}
	return account, nil
}

func testAccount() entity.Account {
	now := time.Now().UTC()
	return entity.Account{ID: "a1", Name: "Cash", Type: entity.TypeAsset, CreatedAt: now, UpdatedAt: now}
}

func TestRetryAccountRepositoryRetriesTransientFailure(t *testing.T) {
	inner := newStubAccountRepository()
	inner.createErrors = []error{errors.New("transient"), nil}
	repo := &RetryAccountRepository{inner: inner, maxAttempts: 3, delay: time.Millisecond, maxDelay: 2 * time.Millisecond}

	if err := repo.Create(context.Background(), "u1", testAccount()); err != nil {
		t.Fatalf("create: %v", err)
	}
	if inner.createCalls != 2 {
		t.Fatalf("expected two attempts, got %d", inner.createCalls)
	}
}

func TestRetryAccountRepositoryDoesNotRetryTerminalError(t *testing.T) {
	inner := newStubAccountRepository()
	inner.createErrors = []error{port.ErrAccountExists}
	repo := &RetryAccountRepository{inner: inner, maxAttempts: 3, delay: time.Millisecond, maxDelay: 2 * time.Millisecond}

	err := repo.Create(context.Background(), "u1", testAccount())
	if !errors.Is(err, port.ErrAccountExists) {
		t.Fatalf("expected ErrAccountExists, got %v", err)
	}
	if inner.createCalls != 1 {
		t.Fatalf("expected one attempt, got %d", inner.createCalls)
	}
}
