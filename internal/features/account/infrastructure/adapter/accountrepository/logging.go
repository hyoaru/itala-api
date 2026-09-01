package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
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

func (c *LoggingAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	logger.Debug("Deleting account", "id", id)

	if err := c.inner.Delete(ctx, userID, id); err != nil {
		logger.Warn("Failed to delete account", "error", err)
		return err
	}

	logger.Info("Account deleted", "id", id)

	return nil
}

func (c *LoggingAccountRepository) AdjustBalance(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
	logger.Debug("Adjusting account balance", "id", accountID, "delta", delta.String())

	if err := c.inner.AdjustBalance(ctx, userID, accountID, idempotencyKey, delta); err != nil {
		logger.Warn("Failed to adjust account balance", "error", err)
		return err
	}

	logger.Info("Account balance adjusted", "id", accountID)

	return nil
}
