package api

import (
	"github.com/hyoaru/itala-api/internal/features/account"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
)

type AccountHandler struct {
	CreateAccount  usecase.UseCase[account.CreateAccountRequest, account.CreateAccountResponse]
	ListAccounts   usecase.UseCase[account.ListAccountsRequest, account.ListAccountsResponse]
	GetAccount     usecase.UseCase[account.GetAccountRequest, account.GetAccountResponse]
	UpdateAccount  usecase.UseCase[account.UpdateAccountRequest, account.UpdateAccountResponse]
	ArchiveAccount usecase.UseCase[account.ArchiveAccountRequest, account.ArchiveAccountResponse]
	RestoreAccount usecase.UseCase[account.RestoreAccountRequest, account.RestoreAccountResponse]
}
