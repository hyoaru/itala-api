package account

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestListAccounts_Execute(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	balance, err := valueobject.NewDecimal("100")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}

	account := entity.Account{
		ID:        "account-1",
		Name:      "Savings",
		Balance:   balance,
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
	}
	nextCursor := "cursor-1"

	t.Run("success", func(t *testing.T) {
		client := &fakeAccountRepository{
			find: func(
				ctx context.Context,
				userID string,
				query accountrepository.AccountQuery,
			) (accountrepository.AccountPage, error) {
				return accountrepository.AccountPage{
					Accounts:   []entity.Account{account},
					NextCursor: &nextCursor,
				}, nil
			},
		}

		useCase := NewListAccounts(client)

		request := ListAccountsRequest{UserID: "user-1", Limit: 10}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		want := ListAccountsResponse{
			Accounts:   []entity.Account{account},
			NextCursor: &nextCursor,
		}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeAccountRepository{
			find: func(
				ctx context.Context,
				userID string,
				query accountrepository.AccountQuery,
			) (accountrepository.AccountPage, error) {
				return accountrepository.AccountPage{}, boom
			},
		}

		useCase := NewListAccounts(client)

		request := ListAccountsRequest{UserID: "user-1", Limit: 10}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if len(got.Accounts) != 0 || got.NextCursor != nil {
			t.Errorf("got %v, want %v", got, ListAccountsResponse{})
		}
	})

	t.Run("writes expected query", func(t *testing.T) {
		var capturedUserID string
		var capturedQuery accountrepository.AccountQuery

		client := &fakeAccountRepository{
			find: func(
				ctx context.Context,
				userID string,
				query accountrepository.AccountQuery,
			) (accountrepository.AccountPage, error) {
				capturedUserID = userID
				capturedQuery = query
				return accountrepository.AccountPage{}, nil
			},
		}

		useCase := NewListAccounts(client)

		name := "Savings"
		cursor := "cursor-1"

		request := ListAccountsRequest{UserID: "user-1", Limit: 10, Name: &name, Cursor: &cursor}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedQuery.Limit != 10 {
			t.Errorf("got %v, want %v", capturedQuery.Limit, 10)
		}

		if capturedQuery.Name == nil || *capturedQuery.Name != "Savings" {
			t.Errorf("got %v, want %v", capturedQuery.Name, "Savings")
		}

		if capturedQuery.Cursor == nil || *capturedQuery.Cursor != "cursor-1" {
			t.Errorf("got %v, want %v", capturedQuery.Cursor, "cursor-1")
		}
	})
}
