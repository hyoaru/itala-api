package transaction

import (
	"context"
	"time"

	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type TransactionQuery struct {
	Type       *valueobjects.TransactionType
	CategoryID *string
	From       *time.Time
	To         *time.Time
}

type TransactionRepository interface {
	Create(ctx context.Context, userID string, transaction entities.Transaction) error
	Find(ctx context.Context, userID string, query TransactionQuery) ([]entities.Transaction, error)
}
