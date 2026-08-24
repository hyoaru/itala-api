package transaction

import (
	category "github.com/hyoaru/itala-api/internal/features/category"
	transactionrepositoryport "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	transactionusecases "github.com/hyoaru/itala-api/internal/features/transaction/application/usecases"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	transactionrepositoryadapters "github.com/hyoaru/itala-api/internal/features/transaction/infrastructure/adapters/transactionrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type Transaction = entities.Transaction

type TransactionRepository = transactionrepositoryport.TransactionRepository

func NewDynamoDBTransactionRepository(client dynamodbclient.DynamoDBClient, tableName string) TransactionRepository {
	r := transactionrepositoryadapters.NewDynamoDBTransactionRepository(client, tableName)
	return transactionrepositoryadapters.NewDecoratedTransactionRepository(r)
}

type CreateTransactionRequest = transactionusecases.CreateTransactionRequest

func NewCreateTransaction(
	transactionRepository TransactionRepository,
	categoryRepository category.CategoryRepository,
) usecases.UseCase[CreateTransactionRequest, struct{}] {
	return transactionusecases.NewCreateTransaction(transactionRepository, categoryRepository)
}

type (
	ListTransactionsRequest  = transactionusecases.ListTransactionsRequest
	ListTransactionsResponse = transactionusecases.ListTransactionsResponse
)

func NewListTransactions(transactionRepository TransactionRepository) usecases.UseCase[ListTransactionsRequest, transactionusecases.ListTransactionsResponse] {
	return transactionusecases.NewListTransactions(transactionRepository)
}
