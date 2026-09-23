package transaction

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestGetTransaction_Execute(t *testing.T) {
	amount, err := valueobject.NewDecimal("10")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	occurredAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	transaction := entity.Transaction{
		ID:          "transaction-1",
		Amount:      amount,
		Type:        valueobject.TransactionTypeExpense,
		AccountID:   "account-1",
		CategoryID:  "category-1",
		Description: "Lunch",
		OccurredAt:  occurredAt,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}

	t.Run("success", func(t *testing.T) {
		client := &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				return transaction, nil
			},
		}

		useCase := NewGetTransaction(client)

		request := GetTransactionRequest{UserID: "user-1", ID: "transaction-1"}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if want := GetTransactionResponse(transaction); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				return entity.Transaction{}, boom
			},
		}

		useCase := NewGetTransaction(client)

		request := GetTransactionRequest{UserID: "user-1", ID: "transaction-1"}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if got != (GetTransactionResponse{}) {
			t.Errorf("got %v, want %v", got, GetTransactionResponse{})
		}
	})

	t.Run("writes expected request", func(t *testing.T) {
		var capturedUserID string
		var capturedID string

		client := &fakeTransactionRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Transaction, error) {
				capturedUserID = userID
				capturedID = id
				return entity.Transaction{}, nil
			},
		}

		useCase := NewGetTransaction(client)

		request := GetTransactionRequest{UserID: "user-1", ID: "transaction-1"}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedID != "transaction-1" {
			t.Errorf("got %v, want %v", capturedID, "transaction-1")
		}
	})
}
