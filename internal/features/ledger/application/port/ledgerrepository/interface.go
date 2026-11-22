package ledgerrepository

import (
	"context"

	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
)

type IdempotencyWrite struct {
	KeyHash       string
	RequestHash   string
	TransactionID string
	CreatedAt     string
}

type RecordTransactionInput struct {
	TransactionID    string
	Description      string
	OccurredAt       string
	CreatedAt        string
	Postings         []entity.Posting
	Summary          entity.TransactionSummary
	StatementEntries []entity.AccountStatementEntry
	BalanceDeltas    []entity.BalanceDelta
	Idempotency      IdempotencyWrite
}

type LedgerRepository interface {
	RecordTransaction(ctx context.Context, userID string, input RecordTransactionInput) error
	FindIdempotency(ctx context.Context, userID string, keyHash string) (entity.IdempotencyRecord, bool, error)
	FindTransactionDetail(ctx context.Context, userID string, occurredAt string, id string) (entity.Transaction, []entity.Posting, error)
}
