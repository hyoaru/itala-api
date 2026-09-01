package account

import (
	"context"

	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type AccountQuery struct {
	Limit  int32
	Name   *string
	Cursor *string
}

type AccountPage struct {
	Accounts   []entity.Account
	NextCursor *string
}

type AccountRepository interface {
	Create(ctx context.Context, userID string, account entity.Account) error
	Find(ctx context.Context, userID string, query AccountQuery) (AccountPage, error)
	FindOne(ctx context.Context, userID string, id string) (entity.Account, error)
	Update(ctx context.Context, userID string, account entity.Account) error
	Delete(ctx context.Context, userID string, id string) error
	AdjustBalance(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error
}
