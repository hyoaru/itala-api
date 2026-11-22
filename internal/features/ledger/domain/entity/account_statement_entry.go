package entity

import "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"

type AccountStatementEntry struct {
	ID                    string
	TransactionID         string
	AccountID             string
	Amount                valueobject.Decimal
	Description           string
	TransactionOccurredAt string
	CounterpartyAccountID string
}
