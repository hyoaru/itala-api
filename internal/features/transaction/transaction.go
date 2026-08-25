package transaction

import (
	category "github.com/hyoaru/itala-api/internal/features/category"
	transactionrepositoryport "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	transactionusecase "github.com/hyoaru/itala-api/internal/features/transaction/application/usecase"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	transactionrepositoryadapter "github.com/hyoaru/itala-api/internal/features/transaction/infrastructure/adapter/transactionrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type Transaction = entity.Transaction

type TransactionRepository = transactionrepositoryport.TransactionRepository

func NewDynamoDBTransactionRepository(client dynamodbclient.DynamoDBClient, tableName string) TransactionRepository {
	r := transactionrepositoryadapter.NewDynamoDBTransactionRepository(client, tableName)
	return transactionrepositoryadapter.NewDecoratedTransactionRepository(r)
}

type (
	CreateTransactionRequest  = transactionusecase.CreateTransactionRequest
	CreateTransactionResponse = transactionusecase.CreateTransactionResponse
)

func NewCreateTransaction(
	transactionRepository TransactionRepository,
	categoryRepository category.CategoryRepository,
) usecase.UseCase[CreateTransactionRequest, CreateTransactionResponse] {
	return transactionusecase.NewCreateTransaction(transactionRepository, categoryRepository)
}

type (
	ListTransactionsRequest  = transactionusecase.ListTransactionsRequest
	ListTransactionsResponse = transactionusecase.ListTransactionsResponse
)

func NewListTransactions(transactionRepository TransactionRepository) usecase.UseCase[ListTransactionsRequest, transactionusecase.ListTransactionsResponse] {
	return transactionusecase.NewListTransactions(transactionRepository)
}

type (
	UpdateTransactionRequest  = transactionusecase.UpdateTransactionRequest
	UpdateTransactionResponse = transactionusecase.UpdateTransactionResponse
)

func NewUpdateTransaction(
	transactionRepository TransactionRepository,
	categoryRepository category.CategoryRepository,
) usecase.UseCase[UpdateTransactionRequest, UpdateTransactionResponse] {
	return transactionusecase.NewUpdateTransaction(transactionRepository, categoryRepository)
}
