package account

import (
	accountrepositoryport "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	accountusecase "github.com/hyoaru/itala-api/internal/features/account/application/usecase"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	accountrepositoryadapter "github.com/hyoaru/itala-api/internal/features/account/infrastructure/adapter/accountrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type Account = entity.Account

type AccountRepository = accountrepositoryport.AccountRepository

var (
	ErrAccountExists   = accountrepositoryport.ErrAccountExists
	ErrAccountNotFound = accountrepositoryport.ErrAccountNotFound
	ErrAccountDeleted  = accountrepositoryport.ErrAccountDeleted
)

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string, idempotencyStore idempotency.IdempotencyStore) AccountRepository {
	r := accountrepositoryadapter.NewDynamoDBAccountRepository(client, tableName)
	return accountrepositoryadapter.NewDecoratedAccountRepository(r, idempotencyStore)
}

type (
	CreateAccountRequest  = accountusecase.CreateAccountRequest
	CreateAccountResponse = accountusecase.CreateAccountResponse
)

func NewCreateAccount(accountRepository AccountRepository) usecase.UseCase[CreateAccountRequest, CreateAccountResponse] {
	return accountusecase.NewCreateAccount(accountRepository)
}

type (
	ListAccountsRequest  = accountusecase.ListAccountsRequest
	ListAccountsResponse = accountusecase.ListAccountsResponse
)

func NewListAccounts(accountRepository AccountRepository) usecase.UseCase[ListAccountsRequest, accountusecase.ListAccountsResponse] {
	return accountusecase.NewListAccounts(accountRepository)
}

type (
	GetAccountRequest  = accountusecase.GetAccountRequest
	GetAccountResponse = accountusecase.GetAccountResponse
)

func NewGetAccount(accountRepository AccountRepository) usecase.UseCase[GetAccountRequest, GetAccountResponse] {
	return accountusecase.NewGetAccount(accountRepository)
}

type (
	UpdateAccountRequest  = accountusecase.UpdateAccountRequest
	UpdateAccountResponse = accountusecase.UpdateAccountResponse
)

func NewUpdateAccount(accountRepository AccountRepository) usecase.UseCase[UpdateAccountRequest, UpdateAccountResponse] {
	return accountusecase.NewUpdateAccount(accountRepository)
}

type (
	DeleteAccountRequest  = accountusecase.DeleteAccountRequest
	DeleteAccountResponse = accountusecase.DeleteAccountResponse
)

func NewDeleteAccount(accountRepository AccountRepository) usecase.UseCase[DeleteAccountRequest, DeleteAccountResponse] {
	return accountusecase.NewDeleteAccount(accountRepository)
}
