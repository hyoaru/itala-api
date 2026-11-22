package account

import (
	"context"
	"errors"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestUpdateAccountRenamesMetadata(t *testing.T) {
	repo := newFakeRepository()
	now := time.Now().UTC()
	if err := repo.Create(context.Background(), "u1", entity.Account{
		ID: "a1", Name: "Cash", Type: entity.TypeAsset, CreatedAt: now, UpdatedAt: now,
	}); err != nil {
		t.Fatalf("seed: %v", err)
	}

	time.Sleep(time.Millisecond)
	updated, err := NewUpdateAccount(repo).Execute(context.Background(), UpdateAccountRequest{
		UserID: "u1", ID: "a1", Name: "Wallet",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Wallet" || !updated.UpdatedAt.After(now) {
		t.Fatalf("unexpected result: %+v", updated)
	}
}

func TestUpdateAccountRejectsEmptyName(t *testing.T) {
	_, err := NewUpdateAccount(newFakeRepository()).Execute(context.Background(), UpdateAccountRequest{
		UserID: "u1", ID: "a1", Name: "",
	})
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("expected ErrInvalidAccount, got %v", err)
	}
}
