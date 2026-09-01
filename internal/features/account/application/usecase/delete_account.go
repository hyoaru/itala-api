package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
)

type DeleteAccountRequest struct {
	UserID string
	ID     string
}

type DeleteAccountResponse struct{}

type DeleteAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewDeleteAccount(accountRepository accountrepository.AccountRepository) *DeleteAccount {
	return &DeleteAccount{accountRepository: accountRepository}
}

func (u *DeleteAccount) Execute(ctx context.Context, request DeleteAccountRequest) (DeleteAccountResponse, error) {
	current, err := u.accountRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return DeleteAccountResponse{}, err
	}

	if current.DeletedAt != nil {
		return DeleteAccountResponse{}, accountrepository.ErrAccountNotFound
	}

	if err := u.accountRepository.Delete(ctx, request.UserID, request.ID); err != nil {
		return DeleteAccountResponse{}, err
	}

	return DeleteAccountResponse{}, nil
}
