package account

import (
	"context"

	port "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

type LoggingAccountRepository struct {
	inner port.AccountRepository
}

func NewLoggingAccountRepository(inner port.AccountRepository) *LoggingAccountRepository {
	return &LoggingAccountRepository{inner: inner}
}

func (c *LoggingAccountRepository) Create(ctx context.Context, userID string, name string) error {
	logger.Debug("Creating account", "name", name)

	if err := c.inner.Create(ctx, userID, name); err != nil {
		logger.Warn("Failed to create account", "error", err)
		return err
	}

	logger.Info("Account created", "name", name)

	return nil
}
