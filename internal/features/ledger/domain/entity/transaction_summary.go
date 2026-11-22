package entity

import "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"

type TransactionSummary struct {
	ID                    string
	TransactionID         string
	Description           string
	OccurredAt            string
	DisplayAmount         valueobject.Decimal
	DisplayAccountID      string
	CounterpartyAccountID string
}
