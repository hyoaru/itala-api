package api

import (
	"net/http"
	"os"

	handler "github.com/hyoaru/itala-api/internal/app/api/handler"
	account "github.com/hyoaru/itala-api/internal/features/account"
	category "github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type App struct{ Handler http.Handler }

func New() *App {
	dynamodbTableName := os.Getenv("DYNAMODB_TABLE_NAME")
	dynamodbClient := dynamodbclient.NewSDKDynamoDBClient()
	idempotencyStore := idempotency.NewDecoratedIdempotencyStore(idempotency.NewDynamoDBIdempotencyStore(dynamodbClient, dynamodbTableName))

	identityProvider := identity.NewCognitoIdentityProvider(os.Getenv("AWS_REGION"), os.Getenv("COGNITO_USER_POOL_ID"))
	categoryRepository := category.NewDynamoDBCategoryRepository(dynamodbClient, dynamodbTableName)
	accountRepository := account.NewDynamoDBAccountRepository(dynamodbClient, dynamodbTableName, idempotencyStore)
	transactionRepository := transaction.NewDynamoDBTransactionRepository(dynamodbClient, dynamodbTableName, idempotencyStore)

	categoryHandler := &handler.CategoryHandler{
		CreateCategory: category.NewCreateCategory(categoryRepository),
		ListCategories: category.NewListCategories(categoryRepository),
		GetCategory:    category.NewGetCategory(categoryRepository),
		UpdateCategory: category.NewUpdateCategory(categoryRepository),
		DeleteCategory: category.NewDeleteCategory(categoryRepository),
	}
	accountHandler := &handler.AccountHandler{
		CreateAccount: account.NewCreateAccount(accountRepository),
		ListAccounts:  account.NewListAccounts(accountRepository),
		GetAccount:    account.NewGetAccount(accountRepository),
		UpdateAccount: account.NewUpdateAccount(accountRepository),
		DeleteAccount: account.NewDeleteAccount(accountRepository),
	}
	transactionHandler := &handler.TransactionHandler{
		CreateTransaction: transaction.NewCreateTransaction(transactionRepository, categoryRepository, accountRepository),
		ListTransactions:  transaction.NewListTransactions(transactionRepository),
		GetTransaction:    transaction.NewGetTransaction(transactionRepository),
		UpdateTransaction: transaction.NewUpdateTransaction(transactionRepository, categoryRepository, accountRepository),
		DeleteTransaction: transaction.NewDeleteTransaction(transactionRepository, accountRepository),
	}

	router := NewRouter(identityProvider, *categoryHandler, *accountHandler, *transactionHandler)

	return &App{Handler: router}
}
