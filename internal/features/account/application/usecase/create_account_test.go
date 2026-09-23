package account

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestCreateAccount_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				return nil
			},
		}

		useCase := NewCreateAccount(client)

		request := CreateAccountRequest{UserID: "user-1", Name: "Savings"}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if _, err := uuid.Parse(got.ID); err != nil {
			t.Errorf("parse id: %v", err)
		}

		if got.Name != "Savings" {
			t.Errorf("got %v, want %v", got.Name, "Savings")
		}

		if got.Balance.String() != "0" {
			t.Errorf("got %v, want %v", got.Balance.String(), "0")
		}

		if got.CreatedAt.IsZero() {
			t.Error("got zero created_at, want non-zero")
		}

		if !got.CreatedAt.Equal(got.UpdatedAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, got.CreatedAt)
		}

		if got.DeletedAt != nil {
			t.Errorf("got %v, want %v", got.DeletedAt, nil)
		}
	})

	t.Run("trims name", func(t *testing.T) {
		client := &fakeAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				return nil
			},
		}

		useCase := NewCreateAccount(client)

		request := CreateAccountRequest{UserID: "user-1", Name: "  Savings  "}

		got, err := useCase.Execute(context.Background(), request)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.Name != "Savings" {
			t.Errorf("got %v, want %v", got.Name, "Savings")
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				return boom
			},
		}

		useCase := NewCreateAccount(client)

		request := CreateAccountRequest{UserID: "user-1", Name: "Savings"}

		got, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}

		if got != (CreateAccountResponse{}) {
			t.Errorf("got %v, want %v", got, CreateAccountResponse{})
		}
	})

	t.Run("writes expected account", func(t *testing.T) {
		var capturedUserID string
		var capturedAccount entity.Account

		client := &fakeAccountRepository{
			create: func(ctx context.Context, userID string, account entity.Account) error {
				capturedUserID = userID
				capturedAccount = account
				return nil
			},
		}

		useCase := NewCreateAccount(client)

		request := CreateAccountRequest{UserID: "user-1", Name: "  Savings  "}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedAccount.ID == "" {
			t.Error("got empty id, want non-empty")
		}

		if capturedAccount.Name != "Savings" {
			t.Errorf("got %v, want %v", capturedAccount.Name, "Savings")
		}

		if capturedAccount.Balance.String() != "0" {
			t.Errorf("got %v, want %v", capturedAccount.Balance.String(), "0")
		}

		if !capturedAccount.CreatedAt.Equal(capturedAccount.UpdatedAt) {
			t.Errorf("got %v, want %v", capturedAccount.UpdatedAt, capturedAccount.CreatedAt)
		}

		if capturedAccount.DeletedAt != nil {
			t.Errorf("got %v, want %v", capturedAccount.DeletedAt, nil)
		}
	})
}
