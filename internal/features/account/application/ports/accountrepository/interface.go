package account

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
)

type AccountQuery struct {
	Limit  int32
	Name   *string
	Cursor *string
}

type AccountPage struct {
	Accounts   []entities.Account
	NextCursor *string
}

type AccountRepository interface {
	Create(ctx context.Context, userID string, account entities.Account) error
	Find(ctx context.Context, userID string, query AccountQuery) (AccountPage, error)
}
