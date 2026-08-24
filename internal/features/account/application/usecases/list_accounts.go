package account

import (
	"context"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
)

type ListAccountsRequest struct {
	UserID string
	Limit  int32
	Name   *string
	Cursor *string
}

type ListAccountsResponse struct {
	Accounts   []entities.Account
	NextCursor *string
}

type ListAccounts struct {
	accountRepository accountrepository.AccountRepository
}

func NewListAccounts(accountRepository accountrepository.AccountRepository) *ListAccounts {
	return &ListAccounts{accountRepository: accountRepository}
}

func (u *ListAccounts) Execute(ctx context.Context, request ListAccountsRequest) (ListAccountsResponse, error) {
	query := accountrepository.AccountQuery{
		Limit:  request.Limit,
		Name:   request.Name,
		Cursor: request.Cursor,
	}

	page, err := u.accountRepository.Find(ctx, request.UserID, query)
	if err != nil {
		return ListAccountsResponse{}, err
	}

	return ListAccountsResponse{Accounts: page.Accounts, NextCursor: page.NextCursor}, nil
}
