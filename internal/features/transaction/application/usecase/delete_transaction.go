package transaction

import (
	"context"

	"github.com/google/uuid"
	account "github.com/hyoaru/itala-api/internal/features/account"
	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type DeleteTransactionRequest struct {
	UserID string
	ID     string
}

type DeleteTransactionResponse struct{}

type DeleteTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
	accountRepository     account.AccountRepository
}

func NewDeleteTransaction(
	transactionRepository transactionrepository.TransactionRepository,
	accountRepository account.AccountRepository,
) *DeleteTransaction {
	return &DeleteTransaction{
		transactionRepository: transactionRepository,
		accountRepository:     accountRepository,
	}
}

func (u *DeleteTransaction) Execute(ctx context.Context, request DeleteTransactionRequest) (DeleteTransactionResponse, error) {
	foundTransaction, err := u.transactionRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return DeleteTransactionResponse{}, err
	}

	if err := u.transactionRepository.Delete(ctx, request.UserID, request.ID); err != nil {
		return DeleteTransactionResponse{}, err
	}

	delta := foundTransaction.Amount
	if foundTransaction.Type == valueobject.TransactionTypeIncome {
		delta = delta.Negate()
	}

	balanceIdempotencyKey := uuid.New().String()
	if err := u.accountRepository.AdjustBalance(ctx, request.UserID, foundTransaction.AccountID, balanceIdempotencyKey, delta); err != nil {
		return DeleteTransactionResponse{}, err
	}

	return DeleteTransactionResponse{}, nil
}
