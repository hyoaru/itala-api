package accountrepository

import (
	"context"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestLoggingAccountRepositoryDelegatesCreate(t *testing.T) {
	inner := newFakeRepository()
	repo := NewLoggingAccountRepository(inner)
	now := time.Now().UTC()

	err := repo.Create(context.Background(), "u1", entity.Account{
		ID: "a1", Name: "Cash", Type: entity.TypeAsset, CreatedAt: now, UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if inner.createCalls != 1 {
		t.Fatalf("expected one delegated create, got %d", inner.createCalls)
	}
	if _, ok := inner.accounts["a1"]; !ok {
		t.Fatal("expected account to be stored")
	}
}
