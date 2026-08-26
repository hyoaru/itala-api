package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingAccountRepository struct {
	inner port.AccountRepository
}

func NewLoggingAccountRepository(inner port.AccountRepository) *LoggingAccountRepository {
	return &LoggingAccountRepository{inner: inner}
}

func (c *LoggingAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	logger.Debug("Creating account", "name", account.Name)

	if err := c.inner.Create(ctx, userID, account); err != nil {
		logger.Warn("Failed to create account", "error", err)
		return err
	}

	logger.Info("Account created", "name", account.Name)

	return nil
}

func (c *LoggingAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	logger.Debug("Finding accounts", "query", query)

	result, err := c.inner.Find(ctx, userID, query)
	if err != nil {
		logger.Warn("Failed to find accounts", "error", err)
		return result, err
	}

	logger.Info("Accounts found", "count", len(result.Accounts))
	return result, nil
}

func (c *LoggingAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	logger.Debug("Finding account", "id", id)

	result, err := c.inner.FindOne(ctx, userID, id)
	if err != nil {
		logger.Warn("Failed to find account", "error", err)
		return result, err
	}

	logger.Info("Account found", "id", id)

	return result, nil
}

func (c *LoggingAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	logger.Debug("Updating account", "id", account.ID)

	if err := c.inner.Update(ctx, userID, account); err != nil {
		logger.Warn("Failed to update account", "error", err)
		return err
	}

	logger.Info("Account updated", "id", account.ID)

	return nil
}

func (c *LoggingAccountRepository) Archive(ctx context.Context, userID string, id string) error {
	logger.Debug("Archiving account", "id", id)

	if err := c.inner.Archive(ctx, userID, id); err != nil {
		logger.Warn("Failed to archive account", "error", err)
		return err
	}

	logger.Info("Account archived", "id", id)

	return nil
}

func (c *LoggingAccountRepository) Restore(ctx context.Context, userID string, id string) error {
	logger.Debug("Restoring account", "id", id)

	if err := c.inner.Restore(ctx, userID, id); err != nil {
		logger.Warn("Failed to restore account", "error", err)
		return err
	}

	logger.Info("Account restored", "id", id)

	return nil
}
