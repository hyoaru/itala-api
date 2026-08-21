package transaction

import (
	"context"
	"time"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/shopspring/decimal"
)

type CreateTransactionRequest struct {
	UserID      string
	Amount      decimal.Decimal
	Type        valueobjects.TransactionType
	CategoryID  string
	Description string
	OccuredAt   time.Time
}

type CreateTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
}

func NewCreateTransaction(transactionRepository transactionrepository.TransactionRepository) *CreateTransaction {
	return &CreateTransaction{transactionRepository: transactionRepository}
}

func (u *CreateTransaction) Execute(ctx context.Context, request CreateTransactionRequest) (struct{}, error) {
	err := u.transactionRepository.Create(
		ctx,
		request.UserID,
		request.Amount,
		request.Type,
		request.CategoryID,
		request.Description,
		request.OccuredAt,
	)
	if err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
