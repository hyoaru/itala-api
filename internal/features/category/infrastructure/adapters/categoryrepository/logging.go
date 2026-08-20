package category

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingCategoryRepository struct {
	inner port.CategoryRepository
}

func NewLoggingCategoryRepository(inner port.CategoryRepository) *LoggingCategoryRepository {
	return &LoggingCategoryRepository{inner: inner}
}

func (c *LoggingCategoryRepository) Create(
	ctx context.Context,
	userID string,
	name string,
	transactionType valueobjects.TransactionType,
) error {
	logger.Debug("Creating category", "name", name, "transaction_type", transactionType)

	if err := c.inner.Create(ctx, userID, name, transactionType); err != nil {
		logger.Warn("Failed to create category", "error", err)
		return err
	}

	logger.Info("Category created", "name", name, "transaction_type", transactionType)

	return nil
}
