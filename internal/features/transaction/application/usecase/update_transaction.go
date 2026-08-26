package transaction

import (
	"context"
	"time"

	account "github.com/hyoaru/itala-api/internal/features/account"
	category "github.com/hyoaru/itala-api/internal/features/category"
	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type UpdateTransactionRequest struct {
	UserID      string
	ID          string
	Amount      valueobject.Decimal
	AccountID   string
	CategoryID  string
	Description string
	OccurredAt  time.Time
}

type UpdateTransactionResponse struct{}

type UpdateTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
	categoryRepository    category.CategoryRepository
	accountRepository     account.AccountRepository
}

func NewUpdateTransaction(
	transactionRepository transactionrepository.TransactionRepository,
	categoryRepository category.CategoryRepository,
	accountRepository account.AccountRepository,
) *UpdateTransaction {
	return &UpdateTransaction{
		transactionRepository: transactionRepository,
		categoryRepository:    categoryRepository,
		accountRepository:     accountRepository,
	}
}

func (u *UpdateTransaction) Execute(ctx context.Context, request UpdateTransactionRequest) (UpdateTransactionResponse, error) {
	now := time.Now().UTC()

	foundCategory, err := u.categoryRepository.FindOne(ctx, request.UserID, request.CategoryID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	if foundCategory.Status == category.StatusArchived {
		return UpdateTransactionResponse{}, category.ErrCategoryArchived
	}

	foundAccount, err := u.accountRepository.FindOne(ctx, request.UserID, request.AccountID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	if foundAccount.Status == account.StatusArchived {
		return UpdateTransactionResponse{}, account.ErrAccountArchived
	}

	transaction := entity.Transaction{
		ID:          request.ID,
		Amount:      request.Amount,
		Type:        foundCategory.TransactionType,
		AccountID:   request.AccountID,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt.UTC(),
		UpdatedAt:   now.UTC(),
	}

	if err := u.transactionRepository.Update(ctx, request.UserID, transaction); err != nil {
		return UpdateTransactionResponse{}, err
	}

	return UpdateTransactionResponse{}, nil
}
