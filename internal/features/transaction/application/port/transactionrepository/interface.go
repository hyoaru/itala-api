package transaction

import (
	"context"
	"time"

	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type TransactionQuery struct {
	Limit      int32
	Type       *valueobject.TransactionType
	AccountID  *string
	CategoryID *string
	From       *time.Time
	To         *time.Time
	Cursor     *string
}

type TransactionPage struct {
	Transactions []entity.Transaction
	NextCursor   *string
}

type TransactionRepository interface {
	Create(ctx context.Context, userID string, transaction entity.Transaction) error
	Find(ctx context.Context, userID string, query TransactionQuery) (TransactionPage, error)
}
