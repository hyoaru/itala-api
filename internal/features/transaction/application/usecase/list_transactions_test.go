package transaction

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestListTransactions_Execute(t *testing.T) {
	amount, err := valueobject.NewDecimal("10")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	transaction := entity.Transaction{
		ID:          "transaction-1",
		Amount:      amount,
		Type:        valueobject.TransactionTypeExpense,
		AccountID:   "account-1",
		CategoryID:  "category-1",
		Description: "Lunch",
		OccurredAt:  createdAt,
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt,
	}
	nextCursor := "cursor-1"

	t.Run("success", func(t *testing.T) {
		client := &fakeTransactionRepository{
			find: func(
				ctx context.Context,
				userID string,
				query transactionrepository.TransactionQuery,
			) (transactionrepository.TransactionPage, error) {
				return transactionrepository.TransactionPage{
					Transactions: []entity.Transaction{transaction},
					NextCursor:   &nextCursor,
				}, nil
			},
		}

		useCase := NewListTransactions(client)

		request := ListTransactionsRequest{UserID: "user-1", Limit: 10}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		want := ListTransactionsResponse{
			Transactions: []entity.Transaction{transaction},
			NextCursor:   &nextCursor,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeTransactionRepository{
			find: func(
				ctx context.Context,
				userID string,
				query transactionrepository.TransactionQuery,
			) (transactionrepository.TransactionPage, error) {
				return transactionrepository.TransactionPage{}, boom
			},
		}

		useCase := NewListTransactions(client)

		request := ListTransactionsRequest{UserID: "user-1", Limit: 10}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if len(got.Transactions) != 0 || got.NextCursor != nil {
			t.Errorf("got %v, want %v", got, ListTransactionsResponse{})
		}
	})

	t.Run("writes expected query", func(t *testing.T) {
		var capturedUserID string
		var capturedQuery transactionrepository.TransactionQuery

		client := &fakeTransactionRepository{
			find: func(
				ctx context.Context,
				userID string,
				query transactionrepository.TransactionQuery,
			) (transactionrepository.TransactionPage, error) {
				capturedUserID = userID
				capturedQuery = query
				return transactionrepository.TransactionPage{}, nil
			},
		}

		useCase := NewListTransactions(client)

		transactionType := valueobject.TransactionTypeExpense
		accountID := "account-1"
		categoryID := "category-1"
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		cursor := "cursor-1"

		request := ListTransactionsRequest{
			UserID:     "user-1",
			Limit:      10,
			Type:       &transactionType,
			AccountID:  &accountID,
			CategoryID: &categoryID,
			From:       &from,
			To:         &to,
			Cursor:     &cursor,
		}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedQuery.Limit != 10 {
			t.Errorf("got %v, want %v", capturedQuery.Limit, 10)
		}

		if capturedQuery.Type == nil || *capturedQuery.Type != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", capturedQuery.Type, valueobject.TransactionTypeExpense)
		}

		if capturedQuery.AccountID == nil || *capturedQuery.AccountID != "account-1" {
			t.Errorf("got %v, want %v", capturedQuery.AccountID, "account-1")
		}

		if capturedQuery.CategoryID == nil || *capturedQuery.CategoryID != "category-1" {
			t.Errorf("got %v, want %v", capturedQuery.CategoryID, "category-1")
		}

		if capturedQuery.From == nil || !capturedQuery.From.Equal(from) {
			t.Errorf("got %v, want %v", capturedQuery.From, from)
		}

		if capturedQuery.To == nil || !capturedQuery.To.Equal(to) {
			t.Errorf("got %v, want %v", capturedQuery.To, to)
		}

		if capturedQuery.Cursor == nil || *capturedQuery.Cursor != "cursor-1" {
			t.Errorf("got %v, want %v", capturedQuery.Cursor, "cursor-1")
		}
	})
}
