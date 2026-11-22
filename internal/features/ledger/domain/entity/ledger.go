package entity

import "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"

type Transaction struct {
	ID                      string
	Description             string
	OccurredAt              string
	CreatedAt               string
	ReversesTransactionID   *string
	ReversedByTransactionID *string
}

type Posting struct {
	ID            string
	TransactionID string
	AccountID     string
	Amount        valueobject.Decimal
}

type BalanceDelta struct {
	AccountID string
	Amount    valueobject.Decimal
}

type TransactionSummary struct {
	ID                    string
	TransactionID         string
	Description           string
	OccurredAt            string
	DisplayAmount         valueobject.Decimal
	DisplayAccountID      string
	CounterpartyAccountID string
}

type AccountStatementEntry struct {
	ID                    string
	TransactionID         string
	AccountID             string
	Amount                valueobject.Decimal
	Description           string
	TransactionOccurredAt string
	CounterpartyAccountID string
}

type IdempotencyRecord struct {
	TransactionID string
	RequestHash   string
	CreatedAt     string
}
