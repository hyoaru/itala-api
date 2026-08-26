package transaction

import (
	"context"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
)

type DeleteTransactionRequest struct {
	UserID string
	ID     string
}

type DeleteTransactionResponse struct{}

type DeleteTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
}

func NewDeleteTransaction(transactionRepository transactionrepository.TransactionRepository) *DeleteTransaction {
	return &DeleteTransaction{transactionRepository: transactionRepository}
}

func (u *DeleteTransaction) Execute(ctx context.Context, request DeleteTransactionRequest) (DeleteTransactionResponse, error) {
	if err := u.transactionRepository.Delete(ctx, request.UserID, request.ID); err != nil {
		return DeleteTransactionResponse{}, err
	}

	return DeleteTransactionResponse{}, nil
}
