package category

import (
	"context"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CategoryRepository interface {
	Create(ctx context.Context, userID string, name string, transactionType valueobjects.TransactionType) error
}
