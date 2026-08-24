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

func (c *LoggingCategoryRepository) Find(ctx context.Context, userID string, query port.CategoryQuery) (port.CategoryPage, error) {
	logger.Debug("Finding categories", "query", query)

	result, err := c.inner.Find(ctx, userID, query)
	if err != nil {
		logger.Warn("Failed to find categories", "error", err)
		return result, err
	}

	logger.Info("Categories found", "count", len(result.Categories))
	return result, nil
}

func (c *LoggingCategoryRepository) FindOne(ctx context.Context, userID string, categoryID string) (entities.Category, error) {
	logger.Debug("Finding category", "category_id", categoryID)

	result, err := c.inner.FindOne(ctx, userID, categoryID)
	if err != nil {
		logger.Warn("Failed to find category", "error", err)
		return result, err
	}

	logger.Info("Category found", "category_id", categoryID)
	return result, nil
}
