package account

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestGetAccount_Execute(t *testing.T) {
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

	t.Run("success", func(t *testing.T) {
		client := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				return account, nil
			},
		}

		useCase := NewGetAccount(client)

		request := GetAccountRequest{UserID: "user-1", ID: "account-1"}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if want := GetAccountResponse(account); got != want {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				return entity.Account{}, boom
			},
		}

		useCase := NewGetAccount(client)

		request := GetAccountRequest{UserID: "user-1", ID: "account-1"}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if got != (GetAccountResponse{}) {
			t.Errorf("got %v, want %v", got, GetAccountResponse{})
		}
	})

	t.Run("writes expected request", func(t *testing.T) {
		var capturedUserID string
		var capturedID string

		client := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				capturedUserID = userID
				capturedID = id
				return entity.Account{}, nil
			},
		}

		useCase := NewGetAccount(client)

		request := GetAccountRequest{UserID: "user-1", ID: "account-1"}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedID != "account-1" {
			t.Errorf("got %v, want %v", capturedID, "account-1")
		}
	})
}
