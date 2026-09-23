package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	accountentity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	categoryentity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestCreateTransaction_Execute(t *testing.T) {
	amount, err := valueobject.NewDecimal("10")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}
	occurredAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

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

	t.Run("success expense", func(t *testing.T) {
		var capturedAccountID string
		var capturedDelta valueobject.Decimal

		transactionRepository := &fakeTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				return nil
			},
		}
		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				capturedAccountID = accountID
				capturedDelta = delta
				return nil
			},
		)

		useCase := NewCreateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		request := CreateTransactionRequest{
			UserID:         "user-1",
			Amount:         amount,
			AccountID:      "account-1",
			CategoryID:     "category-1",
			Description:    "  Lunch  ",
			OccurredAt:     occurredAt,
			IdempotencyKey: "idem-1",
		}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if _, err := uuid.Parse(got.ID); err != nil {
			t.Errorf("parse id: %v", err)
		}

		if got.Amount.String() != "10" {
			t.Errorf("got %v, want %v", got.Amount.String(), "10")
		}

		if got.Type != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.Type, valueobject.TransactionTypeExpense)
		}

		if got.AccountID != "account-1" {
			t.Errorf("got %v, want %v", got.AccountID, "account-1")
		}

		if got.CategoryID != "category-1" {
			t.Errorf("got %v, want %v", got.CategoryID, "category-1")
		}

		if got.Description != "Lunch" {
			t.Errorf("got %v, want %v", got.Description, "Lunch")
		}

		if !got.OccurredAt.Equal(occurredAt) {
			t.Errorf("got %v, want %v", got.OccurredAt, occurredAt)
		}

		if got.CreatedAt.IsZero() || !got.CreatedAt.Equal(got.UpdatedAt) {
			t.Errorf("got created_at %v updated_at %v, want equal non-zero", got.CreatedAt, got.UpdatedAt)
		}

		if capturedAccountID != "account-1" {
			t.Errorf("got %v, want %v", capturedAccountID, "account-1")
		}

		if capturedDelta.String() != "-10" {
			t.Errorf("got %v, want %v", capturedDelta.String(), "-10")
		}
	})

	t.Run("success income", func(t *testing.T) {
		var capturedDelta valueobject.Decimal

		transactionRepository := &fakeTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				return nil
			},
		}
		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				capturedDelta = delta
				return nil
			},
		)

		useCase := NewCreateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeIncome), accountRepository)

		request := CreateTransactionRequest{
			UserID:     "user-1",
			Amount:     amount,
			AccountID:  "account-1",
			CategoryID: "category-1",
			OccurredAt: occurredAt,
		}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.Type != valueobject.TransactionTypeIncome {
			t.Errorf("got %v, want %v", got.Type, valueobject.TransactionTypeIncome)
		}

		if capturedDelta.String() != "10" {
			t.Errorf("got %v, want %v", capturedDelta.String(), "10")
		}
	})

	t.Run("category error", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := &fakeTransactionRepository{}
		categoryRepository := &fakeCategoryRepository{
			findOne: func(ctx context.Context, userID string, categoryID string) (categoryentity.Category, error) {
				return categoryentity.Category{}, boom
			},
		}

		useCase := NewCreateTransaction(transactionRepository, categoryRepository, &fakeAccountRepository{})

		_, err := useCase.Execute(context.Background(), CreateTransactionRequest{})

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("account error", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := &fakeTransactionRepository{}
		accountRepository := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (accountentity.Account, error) {
				return accountentity.Account{}, boom
			},
		}

		useCase := NewCreateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), CreateTransactionRequest{})

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := &fakeTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				return boom
			},
		}

		useCase := NewCreateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), newAccountRepository(nil))

		_, err := useCase.Execute(context.Background(), CreateTransactionRequest{})

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("adjust balance error", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := &fakeTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				return nil
			},
		}
		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				return boom
			},
		)

		useCase := NewCreateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), CreateTransactionRequest{})

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected transaction", func(t *testing.T) {
		var capturedUserID string
		var capturedTransaction entity.Transaction
		var capturedIdempotencyKey string

		transactionRepository := &fakeTransactionRepository{
			create: func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error {
				capturedUserID = userID
				capturedTransaction = transaction
				capturedIdempotencyKey = idempotencyKey
				return nil
			},
		}

		useCase := NewCreateTransaction(transactionRepository, newCategoryRepository(valueobject.TransactionTypeExpense), newAccountRepository(nil))

		request := CreateTransactionRequest{
			UserID:         "user-1",
			Amount:         amount,
			AccountID:      "account-1",
			CategoryID:     "category-1",
			Description:    "  Lunch  ",
			OccurredAt:     occurredAt,
			IdempotencyKey: "idem-1",
		}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedTransaction.ID == "" {
			t.Error("got empty id, want non-empty")
		}

		if capturedTransaction.Amount.String() != "10" {
			t.Errorf("got %v, want %v", capturedTransaction.Amount.String(), "10")
		}

		if capturedTransaction.Type != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", capturedTransaction.Type, valueobject.TransactionTypeExpense)
		}

		if capturedTransaction.AccountID != "account-1" {
			t.Errorf("got %v, want %v", capturedTransaction.AccountID, "account-1")
		}

		if capturedTransaction.CategoryID != "category-1" {
			t.Errorf("got %v, want %v", capturedTransaction.CategoryID, "category-1")
		}

		if capturedTransaction.Description != "Lunch" {
			t.Errorf("got %v, want %v", capturedTransaction.Description, "Lunch")
		}

		if !capturedTransaction.OccurredAt.Equal(occurredAt) {
			t.Errorf("got %v, want %v", capturedTransaction.OccurredAt, occurredAt)
		}

		if capturedTransaction.CreatedAt.IsZero() || !capturedTransaction.CreatedAt.Equal(capturedTransaction.UpdatedAt) {
			t.Errorf("got created_at %v updated_at %v, want equal non-zero", capturedTransaction.CreatedAt, capturedTransaction.UpdatedAt)
		}

		if capturedIdempotencyKey != "idem-1" {
			t.Errorf("got %v, want %v", capturedIdempotencyKey, "idem-1")
		}
	})
}
