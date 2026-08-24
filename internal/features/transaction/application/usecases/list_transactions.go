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
	Limit      int32
	Type       *valueobjects.TransactionType
	CategoryID *string
	From       *time.Time
	To         *time.Time
	Cursor     *string
}

type ListTransactionsResponse struct {
	Transactions []entities.Transaction
	NextCursor   *string
}

type ListTransactions struct {
	transactionRepository transactionrepository.TransactionRepository
}

func NewListTransactions(transactionRepository transactionrepository.TransactionRepository) *ListTransactions {
	return &ListTransactions{transactionRepository: transactionRepository}
}

func (u *ListTransactions) Execute(ctx context.Context, request ListTransactionsRequest) (ListTransactionsResponse, error) {
	query := transactionrepository.TransactionQuery{
		Limit:      request.Limit,
		Type:       request.Type,
		CategoryID: request.CategoryID,
		From:       request.From,
		To:         request.To,
		Cursor:     request.Cursor,
	}

	page, err := u.transactionRepository.Find(ctx, request.UserID, query)
	if err != nil {
		return ListTransactionsResponse{}, err
	}

	return ListTransactionsResponse{Transactions: page.Transactions, NextCursor: page.NextCursor}, nil
}
