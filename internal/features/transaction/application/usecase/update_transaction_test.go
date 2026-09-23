package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	accountentity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	categoryentity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type adjustCall struct {
	accountID string
	delta     valueobject.Decimal
}

func TestUpdateTransaction_Execute(t *testing.T) {
	amount, err := valueobject.NewDecimal("10")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}
	occurredAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	existing := entity.Transaction{
		ID:         "transaction-1",
		Amount:     amount,
		Type:       valueobject.TransactionTypeExpense,
		AccountID:  "account-old",
		CategoryID: "category-old",
	}

	newCategoryRepository := func(transactionType valueobject.TransactionType) *fakeCategoryRepository {
		return &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (categoryentity.Category, error) {
				return categoryentity.Category{TransactionType: transactionType}, nil
			},
		}
	}

	newAccountRepository := func(adjustBalance func(context.Context, string, string, string, valueobject.Decimal) error) *fakeAccountRepository {
		if adjustBalance == nil {
			adjustBalance = func(context.Context, string, string, string, valueobject.Decimal) error {
				return nil
			}
		}
		return &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (accountentity.Account, error) {
				return accountentity.Account{ID: id}, nil
			},
			adjustBalance: adjustBalance,
		}
	}

	newTransactionRepository := func() *fakeTransactionRepository {
		return &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				return existing, nil
			},
			update: func(ctx context.Context, userID string, transaction entity.Transaction) error {
				return nil
			},
		}
	}

	request := UpdateTransactionRequest{
		UserID:     "user-1",
		ID:         "transaction-1",
		Amount:     amount,
		AccountID:  "account-new",
		CategoryID: "category-new",
		OccurredAt: occurredAt,
	}

	t.Run("success", func(t *testing.T) {
		var calls []adjustCall

		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				calls = append(calls, adjustCall{accountID: accountID, delta: delta})
				return nil
			},
		)

		useCase := NewUpdateTransaction(newTransactionRepository(), newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if len(calls) != 2 {
			t.Fatalf("got %d adjust calls, want %d", len(calls), 2)
		}

		if calls[0].accountID != "account-old" || calls[0].delta.String() != "10" {
			t.Errorf("got %v, want %v", calls[0], adjustCall{accountID: "account-old", delta: amount})
		}

		if calls[1].accountID != "account-new" || calls[1].delta.String() != "-10" {
			t.Errorf("got %v, want account-new/-10", calls[1])
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				return entity.Transaction{}, boom
			},
		}

		useCase := NewUpdateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), &fakeAccountRepository{})

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("category error", func(t *testing.T) {
		boom := errors.New("boom")

		categoryRepository := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (categoryentity.Category, error) {
				return categoryentity.Category{}, boom
			},
		}

		useCase := NewUpdateTransaction(newTransactionRepository(), categoryRepository, &fakeAccountRepository{})

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("old account error", func(t *testing.T) {
		boom := errors.New("boom")

		accountRepository := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (accountentity.Account, error) {
				return accountentity.Account{}, boom
			},
		}

		useCase := NewUpdateTransaction(newTransactionRepository(), newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("new account error", func(t *testing.T) {
		boom := errors.New("boom")

		accountRepository := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (accountentity.Account, error) {
				if id == "account-new" {
					return accountentity.Account{}, boom
				}
				return accountentity.Account{ID: id}, nil
			},
		}

		useCase := NewUpdateTransaction(newTransactionRepository(), newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("reverse adjust error", func(t *testing.T) {
		boom := errors.New("boom")

		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				return boom
			},
		)

		useCase := NewUpdateTransaction(newTransactionRepository(), newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("update error compensates", func(t *testing.T) {
		boom := errors.New("boom")
		var calls []adjustCall

		transactionRepository := newTransactionRepository()
		transactionRepository.update = func(ctx context.Context, userID string, transaction entity.Transaction) error {
			return boom
		}
		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				calls = append(calls, adjustCall{accountID: accountID, delta: delta})
				return nil
			},
		)

		useCase := NewUpdateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if len(calls) != 2 {
			t.Fatalf("got %d adjust calls, want %d", len(calls), 2)
		}

		if calls[1].accountID != "account-old" || calls[1].delta.String() != "10" {
			t.Errorf("got %v, want account-old/10", calls[1])
		}
	})

	t.Run("forward adjust error", func(t *testing.T) {
		boom := errors.New("boom")
		adjustCalls := 0

		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				adjustCalls++
				if adjustCalls == 2 {
					return boom
				}
				return nil
			},
		)

		useCase := NewUpdateTransaction(newTransactionRepository(), newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected transaction", func(t *testing.T) {
		var capturedTransaction entity.Transaction

		transactionRepository := newTransactionRepository()
		transactionRepository.update = func(ctx context.Context, userID string, transaction entity.Transaction) error {
			capturedTransaction = transaction
			return nil
		}

		useCase := NewUpdateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeIncome), newAccountRepository(nil))

		updateRequest := UpdateTransactionRequest{
			UserID:      "user-1",
			ID:          "transaction-1",
			Amount:      amount,
			AccountID:   "account-new",
			CategoryID:  "category-new",
			Description: "  Dinner  ",
			OccurredAt:  occurredAt,
		}

		_, _ = useCase.Execute(context.Background(), updateRequest)

		if capturedTransaction.ID != "transaction-1" {
			t.Errorf("got %v, want %v", capturedTransaction.ID, "transaction-1")
		}

		if capturedTransaction.Type != valueobject.TransactionTypeIncome {
			t.Errorf("got %v, want %v", capturedTransaction.Type, valueobject.TransactionTypeIncome)
		}

		if capturedTransaction.AccountID != "account-new" {
			t.Errorf("got %v, want %v", capturedTransaction.AccountID, "account-new")
		}

		if capturedTransaction.CategoryID != "category-new" {
			t.Errorf("got %v, want %v", capturedTransaction.CategoryID, "category-new")
		}

		if capturedTransaction.Description != "Dinner" {
			t.Errorf("got %v, want %v", capturedTransaction.Description, "Dinner")
		}

		if !capturedTransaction.OccurredAt.Equal(occurredAt) {
			t.Errorf("got %v, want %v", capturedTransaction.OccurredAt, occurredAt)
		}

		if capturedTransaction.UpdatedAt.IsZero() {
			t.Error("got zero updated_at, want non-zero")
		}
	})
}
