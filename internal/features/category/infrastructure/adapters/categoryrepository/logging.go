package category

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingCategoryRepository struct {
	inner port.CategoryRepository
}

func NewLoggingCategoryRepository(inner port.CategoryRepository) *LoggingCategoryRepository {
	return &LoggingCategoryRepository{inner: inner}
}

func (c *LoggingCategoryRepository) Create(ctx context.Context, userID string, category entities.Category) error {
	logger.Debug("Creating category", "name", category.Name, "transaction_type", category.Type)

	if err := c.inner.Create(ctx, userID, category); err != nil {
		logger.Warn("Failed to create category", "error", err)
		return err
	}

	logger.Info("Category created", "name", category.Name, "transaction_type", category.Type)

	return nil
}
