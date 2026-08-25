package transaction

import (
	"context"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
)

type GetTransactionRequest struct {
	UserID string
	ID     string
}

type GetTransactionResponse entity.Transaction

type GetTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
}

func NewGetTransaction(transactionRepository transactionrepository.TransactionRepository) *GetTransaction {
	return &GetTransaction{transactionRepository: transactionRepository}
}

func (u *GetTransaction) Execute(ctx context.Context, request GetTransactionRequest) (GetTransactionResponse, error) {
	transaction, err := u.transactionRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return GetTransactionResponse{}, err
	}

	return GetTransactionResponse(transaction), nil
}
