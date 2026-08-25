package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type GetAccountRequest struct {
	UserID string
	ID     string
}

type GetAccountResponse entity.Account

type GetAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewGetAccount(accountRepository accountrepository.AccountRepository) *GetAccount {
	return &GetAccount{accountRepository: accountRepository}
}

func (u *GetAccount) Execute(ctx context.Context, request GetAccountRequest) (GetAccountResponse, error) {
	account, err := u.accountRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return GetAccountResponse{}, err
	}

	return GetAccountResponse(account), nil
}
