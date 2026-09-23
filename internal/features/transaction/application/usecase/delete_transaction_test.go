package transaction

import (
	"context"
	"errors"
	"testing"

	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestDeleteTransaction_Execute(t *testing.T) {
	amount, err := valueobject.NewDecimal("10")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}

	newTransactionRepository := func(transactionType valueobject.TransactionType) *fakeTransactionRepository {
		return &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				return entity.Transaction{
					ID:        "transaction-1",
					Amount:    amount,
					Type:      transactionType,
					AccountID: "account-1",
				}, nil
			},
			delete: func(ctx context.Context, userID string, id string) error {
				return nil
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
			adjustBalance: adjustBalance,
		}
	}

	request := DeleteTransactionRequest{UserID: "user-1", ID: "transaction-1"}

	t.Run("success expense", func(t *testing.T) {
		var capturedAccountID string
		var capturedDelta valueobject.Decimal

		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				capturedAccountID = accountID
				capturedDelta = delta
				return nil
			},
		)

		useCase := NewDeleteTransaction(newTransactionRepository(valueobject.TransactionTypeExpense), accountRepository)

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if capturedAccountID != "account-1" {
			t.Errorf("got %v, want %v", capturedAccountID, "account-1")
		}

		if capturedDelta.String() != "10" {
			t.Errorf("got %v, want %v", capturedDelta.String(), "10")
		}
	})

	t.Run("success income", func(t *testing.T) {
		var capturedDelta valueobject.Decimal

		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				capturedDelta = delta
				return nil
			},
		)

		useCase := NewDeleteTransaction(newTransactionRepository(valueobject.TransactionTypeIncome), accountRepository)

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}

		if capturedDelta.String() != "-10" {
			t.Errorf("got %v, want %v", capturedDelta.String(), "-10")
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				return entity.Transaction{}, boom
			},
		}

		useCase := NewDeleteTransaction(transactionRepository, &fakeAccountRepository{})

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		transactionRepository := newTransactionRepository(valueobject.TransactionTypeExpense)
		transactionRepository.delete = func(ctx context.Context, userID string, id string) error {
			return boom
		}

		useCase := NewDeleteTransaction(transactionRepository, &fakeAccountRepository{})

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("adjust balance error", func(t *testing.T) {
		boom := errors.New("boom")

		accountRepository := newAccountRepository(
			func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
				return boom
			},
		)

		useCase := NewDeleteTransaction(newTransactionRepository(valueobject.TransactionTypeExpense), accountRepository)

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected request", func(t *testing.T) {
		var capturedFindUserID string
		var capturedFindID string
		var capturedDeleteUserID string
		var capturedDeleteID string

		transactionRepository := newTransactionRepository(valueobject.TransactionTypeExpense)
		transactionRepository.findOne = func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
			capturedFindUserID = userID
			capturedFindID = id
			return entity.Transaction{ID: id, Amount: amount, Type: valueobject.TransactionTypeExpense, AccountID: "account-1"}, nil
		}
		transactionRepository.delete = func(ctx context.Context, userID string, id string) error {
			capturedDeleteUserID = userID
			capturedDeleteID = id
			return nil
		}

		useCase := NewDeleteTransaction(transactionRepository, newAccountRepository(nil))

		_, _ = useCase.Execute(context.Background(), request)

		if capturedFindUserID != "user-1" || capturedFindID != "transaction-1" {
			t.Errorf("got find %v/%v, want user-1/transaction-1", capturedFindUserID, capturedFindID)
		}

		if capturedDeleteUserID != "user-1" || capturedDeleteID != "transaction-1" {
			t.Errorf("got delete %v/%v, want user-1/transaction-1", capturedDeleteUserID, capturedDeleteID)
		}
	})
}
