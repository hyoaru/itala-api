package account

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

func TestUpdateAccount_Execute(t *testing.T) {
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
			update: func(ctx context.Context, userID string, account entity.Account) error {
				return nil
			},
		}

		useCase := NewUpdateAccount(client)

		request := UpdateAccountRequest{UserID: "user-1", ID: "account-1", Name: "Savings"}

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
	})

	t.Run("get current error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				return entity.Account{}, boom
			},
		}

		useCase := NewUpdateAccount(client)

		request := UpdateAccountRequest{UserID: "user-1", ID: "account-1", Name: "Savings"}

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				return account, nil
			},
			update: func(ctx context.Context, userID string, account entity.Account) error {
				return boom
			},
		}

		useCase := NewUpdateAccount(client)

		request := UpdateAccountRequest{UserID: "user-1", ID: "account-1", Name: "Savings"}

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected account", func(t *testing.T) {
		var capturedUserID string
		var capturedAccount entity.Account

		client := &fakeAccountRepository{
			findOne: func(ctx context.Context, userID string, id string) (entity.Account, error) {
				return account, nil
			},
			update: func(ctx context.Context, userID string, account entity.Account) error {
				capturedUserID = userID
				capturedAccount = account
				return nil
			},
		}

		useCase := NewUpdateAccount(client)

		request := UpdateAccountRequest{UserID: "user-1", ID: "account-1", Name: "  Savings  "}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedAccount.ID != "account-1" {
			t.Errorf("got %v, want %v", capturedAccount.ID, "account-1")
		}

		if capturedAccount.Name != "Savings" {
			t.Errorf("got %v, want %v", capturedAccount.Name, "Savings")
		}

		if capturedAccount.UpdatedAt.IsZero() {
			t.Error("got zero updated_at, want non-zero")
		}
	})
}
