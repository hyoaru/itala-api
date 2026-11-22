package account

import (
	"context"
	"testing"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

func TestListAccountsReturnsSeededAccounts(t *testing.T) {
	repo := newFakeRepository()
	now := time.Now().UTC()
	for _, item := range []entity.Account{
		{ID: "a1", Name: "Cash", Type: entity.TypeAsset, CreatedAt: now, UpdatedAt: now},
		{ID: "a2", Name: "Groceries", Type: entity.TypeExpense, CreatedAt: now, UpdatedAt: now},
	} {
		if err := repo.Create(context.Background(), "u1", item); err != nil {
			t.Fatalf("seed: %v", err)
		}
	}

	page, err := NewListAccounts(repo).Execute(context.Background(), ListAccountsRequest{UserID: "u1", Limit: 40})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Accounts) != 2 {
		t.Fatalf("expected two accounts, got %d", len(page.Accounts))
	}
}
