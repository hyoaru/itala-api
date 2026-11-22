package account

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestGetAccountReturnsStoredMetadata(t *testing.T) {
	repo := newFakeRepository()
	now := time.Now().UTC()
	_ = repo.Create(context.Background(), "u1", entity.Account{
		ID:        "a1",
		Name:      "Cash",
		Type:      entity.TypeAsset,
		CreatedAt: now,
		UpdatedAt: now,
	})

	result, err := NewGetAccount(repo).Execute(context.Background(), GetAccountRequest{UserID: "u1", ID: "a1"})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if result.Name != "Cash" || result.Type != entity.TypeAsset {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestGetAccountMissing(t *testing.T) {
	_, err := NewGetAccount(newFakeRepository()).Execute(context.Background(), GetAccountRequest{UserID: "u1", ID: "missing"})
	if !errors.Is(err, port.ErrAccountNotFound) {
		t.Fatalf("expected ErrAccountNotFound, got %v", err)
	}
}
