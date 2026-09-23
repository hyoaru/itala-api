package account

import (
	"context"
	"errors"
	"testing"
)

func TestDeleteAccount_Execute(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeAccountRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				return nil
			},
		}

		useCase := NewDeleteAccount(client)

		request := DeleteAccountRequest{UserID: "user-1", ID: "account-1"}

		if _, err := useCase.Execute(context.Background(), request); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeAccountRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				return boom
			},
		}

		useCase := NewDeleteAccount(client)

		request := DeleteAccountRequest{UserID: "user-1", ID: "account-1"}

		_, err := useCase.Execute(context.Background(), request)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected request", func(t *testing.T) {
		var capturedUserID string
		var capturedID string

		client := &fakeAccountRepository{
			delete: func(ctx context.Context, userID string, id string) error {
				capturedUserID = userID
				capturedID = id
				return nil
			},
		}

		useCase := NewDeleteAccount(client)

		request := DeleteAccountRequest{UserID: "user-1", ID: "account-1"}

		_, _ = useCase.Execute(context.Background(), request)

		if capturedUserID != "user-1" {
			t.Errorf("got %v, want %v", capturedUserID, "user-1")
		}

		if capturedID != "account-1" {
			t.Errorf("got %v, want %v", capturedID, "account-1")
		}
	})
}
