package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
)

type ArchiveAccountRequest struct {
	UserID string
	ID     string
}

type ArchiveAccountResponse struct{}

type ArchiveAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewArchiveAccount(accountRepository accountrepository.AccountRepository) *ArchiveAccount {
	return &ArchiveAccount{accountRepository: accountRepository}
}

func (u *ArchiveAccount) Execute(ctx context.Context, request ArchiveAccountRequest) (ArchiveAccountResponse, error) {
	if err := u.accountRepository.Archive(ctx, request.UserID, request.ID); err != nil {
		return ArchiveAccountResponse{}, err
	}

	return ArchiveAccountResponse{}, nil
}
