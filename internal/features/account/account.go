package account

import (
	accountrepositoryport "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	accountusecase "github.com/hyoaru/itala-api/internal/features/account/application/usecase"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/features/account/domain/valueobject"
	accountrepositoryadapter "github.com/hyoaru/itala-api/internal/features/account/infrastructure/adapter/accountrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type (
	Account = entity.Account
	Status  = valueobject.Status
)

const (
	StatusActive   = valueobject.StatusActive
	StatusArchived = valueobject.StatusArchived
)

type AccountRepository = accountrepositoryport.AccountRepository

var (
	ErrAccountExists   = accountrepositoryport.ErrAccountExists
	ErrAccountNotFound = accountrepositoryport.ErrAccountNotFound
	ErrAccountArchived = accountrepositoryport.ErrAccountArchived
)

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string) AccountRepository {
	r := accountrepositoryadapter.NewDynamoDBAccountRepository(client, tableName)
	return accountrepositoryadapter.NewDecoratedAccountRepository(r)
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
	UpdateAccountRequest  = accountusecase.UpdateAccountRequest
	UpdateAccountResponse = accountusecase.UpdateAccountResponse
)

func NewUpdateAccount(accountRepository AccountRepository) usecase.UseCase[UpdateAccountRequest, UpdateAccountResponse] {
	return accountusecase.NewUpdateAccount(accountRepository)
}

type (
	GetAccountRequest  = accountusecase.GetAccountRequest
	GetAccountResponse = accountusecase.GetAccountResponse
)

func NewGetAccount(accountRepository AccountRepository) usecase.UseCase[GetAccountRequest, GetAccountResponse] {
	return accountusecase.NewGetAccount(accountRepository)
}
