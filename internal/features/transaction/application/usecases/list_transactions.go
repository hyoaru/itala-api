package transaction

import (
	"context"
	"time"

	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type ListTransactionsRequest struct {
	UserID     string
	Type       *valueobjects.TransactionType
	CategoryID *string
	From       *time.Time
	To         *time.Time
}

type ListTransactions struct {
	transactionRepository transactionrepository.TransactionRepository
}

func NewListTransactions(transactionRepository transactionrepository.TransactionRepository) *ListTransactions {
	return &ListTransactions{transactionRepository: transactionRepository}
}

func (u *ListTransactions) Execute(ctx context.Context, request ListTransactionsRequest) ([]entities.Transaction, error) {
	query := transactionrepository.TransactionQuery{
		Type:       request.Type,
		CategoryID: request.CategoryID,
		From:       request.From,
		To:         request.To,
	}

	transactions, err := u.transactionRepository.Find(ctx, request.UserID, query)
	if err != nil {
		return nil, err
	}

	return transactions, nil
}
