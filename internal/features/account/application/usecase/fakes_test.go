package account

import (
	"context"

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
