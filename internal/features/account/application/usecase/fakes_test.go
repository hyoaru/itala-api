package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type fakeAccountRepository struct {
	accountrepository.AccountRepository
	create        func(ctx context.Context, userID string, account entity.Account) error
	find          func(ctx context.Context, userID string, query accountrepository.AccountQuery) (accountrepository.AccountPage, error)
	findOne       func(ctx context.Context, userID string, id string) (entity.Account, error)
	update        func(ctx context.Context, userID string, account entity.Account) error
	delete        func(ctx context.Context, userID string, id string) error
	adjustBalance func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error
}

func (f *fakeAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	return f.create(ctx, userID, account)
}

func (f *fakeAccountRepository) Find(
	ctx context.Context,
	userID string,
	query accountrepository.AccountQuery,
) (accountrepository.AccountPage, error) {
	return f.find(ctx, userID, query)
}

func (f *fakeAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	return f.findOne(ctx, userID, id)
}

func (f *fakeAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	return f.update(ctx, userID, account)
}

func (f *fakeAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	return f.delete(ctx, userID, id)
}

func (f *fakeAccountRepository) AdjustBalance(
	ctx context.Context,
	userID string,
	accountID string,
	idempotencyKey string,
	delta valueobject.Decimal,
) error {
	return f.adjustBalance(ctx, userID, accountID, idempotencyKey, delta)
}
