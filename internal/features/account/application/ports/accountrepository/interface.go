package account

import "context"

type AccountRepository interface {
	Create(ctx context.Context, userID string, name string) error
}
