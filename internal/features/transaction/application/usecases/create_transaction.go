package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"
	category "github.com/hyoaru/itala-api/internal/features/category"
	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CreateTransactionRequest struct {
	UserID      string
	Amount      valueobjects.Decimal
	AccountID   string
	CategoryID  string
	Description string
	OccurredAt  time.Time
}

type CreateTransactionResponse entities.Transaction

type CreateTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
	categoryRepository    category.CategoryRepository
}

func NewCreateTransaction(
	transactionRepository transactionrepository.TransactionRepository,
	categoryRepository category.CategoryRepository,
) *CreateTransaction {
	return &CreateTransaction{
		transactionRepository: transactionRepository,
		categoryRepository:    categoryRepository,
	}
}

func (u *CreateTransaction) Execute(ctx context.Context, request CreateTransactionRequest) (CreateTransactionResponse, error) {
	id := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	category, err := u.categoryRepository.FindOne(ctx, request.UserID, request.CategoryID)
	if err != nil {
		return CreateTransactionResponse{}, err
	}

	transaction := entities.Transaction{
		ID:          id.String(),
		Amount:      request.Amount,
		Type:        category.TransactionType,
		AccountID:   request.AccountID,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.transactionRepository.Create(ctx, request.UserID, transaction); err != nil {
		return CreateTransactionResponse{}, err
	}

	return CreateTransactionResponse(transaction), nil
}
