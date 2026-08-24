package transaction

import (
	"context"
	"time"

	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type TransactionQuery struct {
	Limit      int32
	Type       *valueobjects.TransactionType
	AccountID  *string
	CategoryID *string
	From       *time.Time
	To         *time.Time
	Cursor     *string
}

type TransactionPage struct {
	Transactions []entities.Transaction
	NextCursor   *string
}

type TransactionRepository interface {
	Create(ctx context.Context, userID string, transaction entities.Transaction) error
	Find(ctx context.Context, userID string, query TransactionQuery) (TransactionPage, error)
}
