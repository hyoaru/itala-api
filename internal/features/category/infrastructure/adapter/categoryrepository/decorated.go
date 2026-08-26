package category

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
)

type DecoratedCategoryRepository struct {
	inner port.CategoryRepository
}

func NewDecoratedCategoryRepository(inner port.CategoryRepository) *DecoratedCategoryRepository {
	return &DecoratedCategoryRepository{inner: NewLoggingCategoryRepository(inner)}
}

func (c *DecoratedCategoryRepository) Create(ctx context.Context, userID string, category entity.Category) error {
	return c.inner.Create(ctx, userID, category)
}

func (c *DecoratedCategoryRepository) Find(ctx context.Context, userID string, query port.CategoryQuery) (port.CategoryPage, error) {
	return c.inner.Find(ctx, userID, query)
}

func (c *DecoratedCategoryRepository) FindOne(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
	return c.inner.FindOne(ctx, userID, categoryID)
}

func (c *DecoratedCategoryRepository) Update(ctx context.Context, userID string, category entity.Category) error {
	return c.inner.Update(ctx, userID, category)
}

func (c *DecoratedCategoryRepository) Archive(ctx context.Context, userID string, categoryID string) error {
	return c.inner.Archive(ctx, userID, categoryID)
}

func (c *DecoratedCategoryRepository) Restore(ctx context.Context, userID string, categoryID string) error {
	return c.inner.Restore(ctx, userID, categoryID)
}
