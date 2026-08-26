package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
)

type RestoreAccountRequest struct {
	UserID string
	ID     string
}

type RestoreAccountResponse struct{}

type RestoreAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewRestoreAccount(accountRepository accountrepository.AccountRepository) *RestoreAccount {
	return &RestoreAccount{accountRepository: accountRepository}
}

func (u *RestoreAccount) Execute(ctx context.Context, request RestoreAccountRequest) (RestoreAccountResponse, error) {
	if err := u.accountRepository.Restore(ctx, request.UserID, request.ID); err != nil {
		return RestoreAccountResponse{}, err
	}

	return RestoreAccountResponse{}, nil
}
