package account

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type fakeRepository struct {
	accounts map[string]entity.Account
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{accounts: map[string]entity.Account{}}
}

func (f *fakeRepository) Create(_ context.Context, _ string, account entity.Account) error {
	f.accounts[account.ID] = account
	return nil
}

func (f *fakeRepository) Find(_ context.Context, _ string, _ port.AccountQuery) (port.AccountPage, error) {
	accounts := make([]entity.Account, 0, len(f.accounts))
	for _, account := range f.accounts {
		accounts = append(accounts, account)
	}
	return port.AccountPage{Accounts: accounts}, nil
}

func (f *fakeRepository) FindOne(_ context.Context, _ string, id string) (entity.Account, error) {
	account, ok := f.accounts[id]
	if !ok {
		return entity.Account{}, port.ErrAccountNotFound
	}
	return account, nil
}

func (f *fakeRepository) Update(_ context.Context, _ string, account entity.Account) error {
	f.accounts[account.ID] = account
	return nil
}

func (f *fakeRepository) Delete(_ context.Context, _ string, id string) error {
	if _, ok := f.accounts[id]; !ok {
		return port.ErrAccountNotFound
	}
	delete(f.accounts, id)
	return nil
}

func TestCreateAccountRejectsInvalidInput(t *testing.T) {
	repo := newFakeRepository()
	useCase := NewCreateAccount(repo)

	_, err := useCase.Execute(context.Background(), CreateAccountRequest{
		UserID: "u1",
		Name:   "",
		Type:   entity.TypeAsset,
	})
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("expected ErrInvalidAccount, got %v", err)
	}

	_, err = useCase.Execute(context.Background(), CreateAccountRequest{
		UserID: "u1",
		Name:   "Cash",
		Type:   entity.Type("investment"),
	})
	if !errors.Is(err, ErrInvalidAccount) {
		t.Fatalf("expected ErrInvalidAccount for invalid type, got %v", err)
	}
}

func TestAccountLifecycle(t *testing.T) {
	repo := newFakeRepository()
	ctx := context.Background()

	created, err := NewCreateAccount(repo).Execute(ctx, CreateAccountRequest{
		UserID: "u1",
		Name:   "Cash",
		Type:   entity.TypeAsset,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if created.ID == "" || created.Type != entity.TypeAsset {
		t.Fatalf("unexpected created account: %+v", created)
	}

	page, err := NewListAccounts(repo).Execute(ctx, ListAccountsRequest{UserID: "u1", Limit: 40})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(page.Accounts) != 1 {
		t.Fatalf("expected one account, got %d", len(page.Accounts))
	}

	before := created.UpdatedAt
	time.Sleep(time.Millisecond)
	updated, err := NewUpdateAccount(repo).Execute(ctx, UpdateAccountRequest{
		UserID: "u1",
		ID:     created.ID,
		Name:   "Wallet",
	})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Wallet" || !updated.UpdatedAt.After(before) {
		t.Fatalf("unexpected updated account: %+v", updated)
	}

	stored, err := NewGetAccount(repo).Execute(ctx, GetAccountRequest{UserID: "u1", ID: created.ID})
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if stored.Name != "Wallet" || stored.Type != entity.TypeAsset {
		t.Fatalf("unexpected stored account: %+v", stored)
	}
}
