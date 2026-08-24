package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CreateTransactionRequest struct {
	UserID      string
	Amount      valueobjects.Decimal
	Type        valueobjects.TransactionType
	CategoryID  string
	Description string
	OccurredAt  time.Time
}

type CreateTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
}

func NewCreateTransaction(transactionRepository transactionrepository.TransactionRepository) *CreateTransaction {
	return &CreateTransaction{transactionRepository: transactionRepository}
}

func (u *CreateTransaction) Execute(ctx context.Context, request CreateTransactionRequest) (struct{}, error) {
	now := time.Now().UTC()

	transaction := entities.Transaction{
		ID:          uuid.New().String(),
		Amount:      request.Amount,
		Type:        request.Type,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.transactionRepository.Create(ctx, request.UserID, transaction); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
