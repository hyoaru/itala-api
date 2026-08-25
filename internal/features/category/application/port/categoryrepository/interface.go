package category

import (
	"context"

	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	categoryvo "github.com/hyoaru/itala-api/internal/features/category/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type CategoryQuery struct {
	Limit           int32
	TransactionType *valueobject.TransactionType
	Name            *string
	Status          *categoryvo.Status
	Cursor          *string
}

type CategoryPage struct {
	Categories []entity.Category
	NextCursor *string
}

type CategoryRepository interface {
	Create(ctx context.Context, userID string, category entity.Category) error
	Find(ctx context.Context, userID string, query CategoryQuery) (CategoryPage, error)
	FindOne(ctx context.Context, userID string, categoryID string) (entity.Category, error)
	Update(ctx context.Context, userID string, category entity.Category) error
}
