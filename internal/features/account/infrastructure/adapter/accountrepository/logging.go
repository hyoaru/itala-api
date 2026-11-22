package accountrepository

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

func (r *LoggingAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	logger.Debug("Creating account", "name", account.Name, "type", account.Type)
	if err := r.inner.Create(ctx, userID, account); err != nil {
		logger.Warn("Failed to create account", "error", err)
		return err
	}
	logger.Info("Account created", "id", account.ID)
	return nil
}

func (r *LoggingAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	page, err := r.inner.Find(ctx, userID, query)
	if err != nil {
		logger.Warn("Failed to find accounts", "error", err)
		return page, err
	}
	logger.Debug("Accounts found", "count", len(page.Accounts))
	return page, nil
}

func (r *LoggingAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	account, err := r.inner.FindOne(ctx, userID, id)
	if err != nil {
		logger.Warn("Failed to find account", "id", id, "error", err)
		return account, err
	}
	return account, nil
}

func (r *LoggingAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	if err := r.inner.Update(ctx, userID, account); err != nil {
		logger.Warn("Failed to update account", "id", account.ID, "error", err)
		return err
	}
	logger.Info("Account updated", "id", account.ID)
	return nil
}

func (r *LoggingAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	if err := r.inner.Delete(ctx, userID, id); err != nil {
		logger.Warn("Failed to delete account", "id", id, "error", err)
		return err
	}
	logger.Info("Account deleted", "id", id)
	return nil
}
