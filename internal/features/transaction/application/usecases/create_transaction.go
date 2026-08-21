package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/shopspring/decimal"
)

type CreateTransactionRequest struct {
	UserID      string
	Amount      decimal.Decimal
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
	id := uuid.New()
	now := time.Now().UTC()
	occurredAt := request.OccurredAt.UTC()

	transaction := entities.Transaction{
		ID:          id.String(),
		Amount:      request.Amount,
		Type:        request.Type,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  occurredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.transactionRepository.Create(ctx, request.UserID, transaction); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
