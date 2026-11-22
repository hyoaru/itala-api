package entity

import "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"

type Posting struct {
	ID            string
	TransactionID string
	AccountID     string
	Amount        valueobject.Decimal
}
