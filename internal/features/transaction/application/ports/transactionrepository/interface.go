package transaction

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
)

type TransactionRepository interface {
	Create(ctx context.Context, userID string, transaction entities.Transaction) error
}
