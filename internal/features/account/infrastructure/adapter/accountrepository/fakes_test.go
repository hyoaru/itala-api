package accountrepository

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type fakeRepository struct {
	accounts     map[string]entity.Account
	createCalls  int
	createErrors []error
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{accounts: map[string]entity.Account{}}
}

func (f *fakeRepository) Create(_ context.Context, _ string, account entity.Account) error {
	f.createCalls++
	if len(f.createErrors) > 0 {
		err := f.createErrors[0]
		f.createErrors = f.createErrors[1:]
		if err != nil {
			return err
		}
	}
	f.accounts[account.ID] = account
	return nil
}

func (f *fakeRepository) Find(_ context.Context, _ string, _ port.AccountQuery) (port.AccountPage, error) {
	return port.AccountPage{}, nil
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
	delete(f.accounts, id)
	return nil
}
