package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CreateAccountRequest struct {
	UserID          string
	Name            string
	TransactionType valueobjects.TransactionType
}

type CreateAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewCreateAccount(accountRepository accountrepository.AccountRepository) *CreateAccount {
	return &CreateAccount{accountRepository: accountRepository}
}

func (u *CreateAccount) Execute(ctx context.Context, request CreateAccountRequest) (struct{}, error) {
	err := u.accountRepository.Create(
		ctx,
		request.UserID,
		request.Name,
	)
	if err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
