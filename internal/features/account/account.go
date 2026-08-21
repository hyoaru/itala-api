package account

import (
	accountrepositoryport "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	accountusecases "github.com/hyoaru/itala-api/internal/features/account/application/usecases"
	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
	accountrepositoryadapters "github.com/hyoaru/itala-api/internal/features/account/infrastructure/adapters/accountrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type Account = entities.Account

type AccountRepository = accountrepositoryport.AccountRepository

var ErrAccountExists = accountrepositoryport.ErrAccountExists

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string) AccountRepository {
	r := accountrepositoryadapters.NewDynamoDBAccountRepository(client, tableName)
	return accountrepositoryadapters.NewDecoratedAccountRepository(r)
}

type CreateAccountRequest = accountusecases.CreateAccountRequest

func NewCreateAccount(accountRepository AccountRepository) usecases.UseCase[CreateAccountRequest, struct{}] {
	return accountusecases.NewCreateAccount(accountRepository)
}
