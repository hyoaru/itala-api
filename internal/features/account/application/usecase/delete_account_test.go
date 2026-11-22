package account

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestDeleteAccountRemovesAccount(t *testing.T) {
	repo := newFakeRepository()
	now := time.Now().UTC()
	if err := repo.Create(context.Background(), "u1", entity.Account{
		ID: "a1", Name: "Cash", Type: entity.TypeAsset, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	if _, err := NewDeleteAccount(repo).Execute(context.Background(), DeleteAccountRequest{UserID: "u1", ID: "a1"}); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := NewGetAccount(repo).Execute(context.Background(), GetAccountRequest{UserID: "u1", ID: "a1"}); !errors.Is(err, port.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}
