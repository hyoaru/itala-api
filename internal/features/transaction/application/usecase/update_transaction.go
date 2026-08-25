package transaction

import (
	"context"
	"time"

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
}

func NewUpdateTransaction(
	transactionRepository transactionrepository.TransactionRepository,
	categoryRepository category.CategoryRepository,
) *UpdateTransaction {
	return &UpdateTransaction{
		transactionRepository: transactionRepository,
		categoryRepository:    categoryRepository,
	}
}

func (u *UpdateTransaction) Execute(ctx context.Context, request UpdateTransactionRequest) (UpdateTransactionResponse, error) {
	now := time.Now().UTC()
	category, err := u.categoryRepository.FindOne(ctx, request.UserID, request.CategoryID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	transaction := entity.Transaction{
		ID:          request.ID,
		Amount:      request.Amount,
		Type:        category.TransactionType,
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
