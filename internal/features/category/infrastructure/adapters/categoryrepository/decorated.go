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
