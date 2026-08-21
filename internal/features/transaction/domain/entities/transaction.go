package transaction

import (
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/shopspring/decimal"
)

type Transaction struct {
	ID          string
	Amount      decimal.Decimal
	Type        valueobjects.TransactionType
	CategoryID  string
	Description string
	OccuredAt   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
