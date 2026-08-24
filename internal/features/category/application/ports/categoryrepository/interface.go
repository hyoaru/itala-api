package category

import (
	"context"

	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CategoryQuery struct {
	Limit  int32
	Type   *valueobjects.TransactionType
	Name   *string
	Cursor *string
}

type CategoryPage struct {
	Categories []entities.Category
	NextCursor *string
}

type CategoryRepository interface {
	Create(ctx context.Context, userID string, category entities.Category) error
	Find(ctx context.Context, userID string, query CategoryQuery) (CategoryPage, error)
}
