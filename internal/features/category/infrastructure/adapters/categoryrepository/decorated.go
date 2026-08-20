package category

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type DecoratedCategoryRepository struct {
	inner port.CategoryRepository
}

func NewDecoratedCategoryRepository(inner port.CategoryRepository) *DecoratedCategoryRepository {
	return &DecoratedCategoryRepository{inner: NewLoggingCategoryRepository(inner)}
}

func (c *DecoratedCategoryRepository) Create(
	ctx context.Context,
	userID string,
	name string,
	transactionType valueobjects.TransactionType,
) error {
	return c.inner.Create(ctx, userID, name, transactionType)
}
