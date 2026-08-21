package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingAccountRepository struct {
	inner port.AccountRepository
}

func NewLoggingAccountRepository(inner port.AccountRepository) *LoggingAccountRepository {
	return &LoggingAccountRepository{inner: inner}
}

func (c *LoggingAccountRepository) Create(ctx context.Context, userID string, account entities.Account) error {
	logger.Debug("Creating account", "name", account.Name)

	if err := c.inner.Create(ctx, userID, account); err != nil {
		logger.Warn("Failed to create account", "error", err)
		return err
	}

	logger.Info("Account created", "name", account.Name)

	return nil
}
