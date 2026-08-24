package category

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
)

type DecoratedCategoryRepository struct {
	inner port.CategoryRepository
}

func NewDecoratedCategoryRepository(inner port.CategoryRepository) *DecoratedCategoryRepository {
	return &DecoratedCategoryRepository{inner: NewLoggingCategoryRepository(inner)}
}

func (c *DecoratedCategoryRepository) Create(ctx context.Context, userID string, category entities.Category) error {
	return c.inner.Create(ctx, userID, category)
}

func (c *DecoratedCategoryRepository) Find(ctx context.Context, userID string, query port.CategoryQuery) (port.CategoryPage, error) {
	return c.inner.Find(ctx, userID, query)
}

func (c *DecoratedCategoryRepository) FindOne(ctx context.Context, userID string, categoryID string) (entities.Category, error) {
	return c.inner.FindOne(ctx, userID, categoryID)
}
