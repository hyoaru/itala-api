package transaction

import (
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type Transaction struct {
	ID          string
	Amount      valueobjects.Decimal
	Type        valueobjects.TransactionType
	AccountID   string
	CategoryID  string
	Description string
	OccurredAt  time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
