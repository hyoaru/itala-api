package transaction

import (
	"context"
	"time"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/shopspring/decimal"
)

type TransactionRepository interface {
	Create(
		ctx context.Context,
		userID string,
		amount decimal.Decimal,
		transactionType valueobjects.TransactionType,
		categoryID string,
		description string,
		occurredAt time.Time,
	) error
}
