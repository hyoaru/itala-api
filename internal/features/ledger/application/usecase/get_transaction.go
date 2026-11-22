package ledger

import (
	"context"

	ledgerrepository "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
)

type GetTransactionRequest struct {
	UserID     string
	OccurredAt string
	ID         string
}

type GetTransactionResponse struct {
	Transaction entity.Transaction
	Postings    []entity.Posting
}

type GetTransaction struct {
	ledgerRepository ledgerrepository.LedgerRepository
}

func NewGetTransaction(ledgerRepository ledgerrepository.LedgerRepository) *GetTransaction {
	return &GetTransaction{ledgerRepository: ledgerRepository}
}

func (u *GetTransaction) Execute(ctx context.Context, request GetTransactionRequest) (GetTransactionResponse, error) {
	transaction, postings, err := u.ledgerRepository.FindTransactionDetail(ctx, request.UserID, request.OccurredAt, request.ID)
	if err != nil {
		return GetTransactionResponse{}, err
	}

	return GetTransactionResponse{Transaction: transaction, Postings: postings}, nil
}
