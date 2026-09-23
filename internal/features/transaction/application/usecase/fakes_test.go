package transaction

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	accountentity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	categoryentity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type fakeTransactionRepository struct {
	transactionrepository.TransactionRepository
	create  func(ctx context.Context, userID string, transaction entity.Transaction, idempotencyKey string) error
	find    func(ctx context.Context, userID string, query transactionrepository.TransactionQuery) (transactionrepository.TransactionPage, error)
	findOne func(ctx context.Context, userID string, id string) (entity.Transaction, error)
	update  func(ctx context.Context, userID string, transaction entity.Transaction) error
	delete  func(ctx context.Context, userID string, id string) error
}

func (f *fakeTransactionRepository) Create(
	ctx context.Context,
	userID string,
	transaction entity.Transaction,
	idempotencyKey string,
) error {
	return f.create(ctx, userID, transaction, idempotencyKey)
}

func (f *fakeTransactionRepository) Find(
	ctx context.Context,
	userID string,
	query transactionrepository.TransactionQuery,
) (transactionrepository.TransactionPage, error) {
	return f.find(ctx, userID, query)
}

func (f *fakeTransactionRepository) FindOne(ctx context.Context, userID string, id string) (entity.Transaction, error) {
	return f.findOne(ctx, userID, id)
}

func (f *fakeTransactionRepository) Update(ctx context.Context, userID string, transaction entity.Transaction) error {
	return f.update(ctx, userID, transaction)
}

func (f *fakeTransactionRepository) Delete(ctx context.Context, userID string, id string) error {
	return f.delete(ctx, userID, id)
}

type fakeCategoryRepository struct {
	categoryrepository.CategoryRepository
	findOne func(ctx context.Context, userID string, categoryID string) (categoryentity.Category, error)
}

func (f *fakeCategoryRepository) FindOne(ctx context.Context, userID string, categoryID string) (categoryentity.Category, error) {
	return f.findOne(ctx, userID, categoryID)
}

type fakeAccountRepository struct {
	accountrepository.AccountRepository
	findOne       func(ctx context.Context, userID string, id string) (accountentity.Account, error)
	adjustBalance func(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error
}

func (f *fakeAccountRepository) FindOne(ctx context.Context, userID string, id string) (accountentity.Account, error) {
	return f.findOne(ctx, userID, id)
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
