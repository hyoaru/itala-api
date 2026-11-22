package api

import (
	"net/http"
	"os"

	handler "github.com/hyoaru/itala-api/internal/app/api/handler"
	account "github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/ledger"
	legacyaccount "github.com/hyoaru/itala-api/internal/legacy/features/account"
	category "github.com/hyoaru/itala-api/internal/legacy/features/category"
	"github.com/hyoaru/itala-api/internal/legacy/features/transaction"
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
	accountRepository := account.NewDynamoDBAccountRepository(dynamodbClient, dynamodbTableName)
	// shortcut: legacy transactions still adjust balance through the legacy repository; remove at cutover.
	legacyAccountRepository := legacyaccount.NewDynamoDBAccountRepository(dynamodbClient, dynamodbTableName, idempotencyStore)
	transactionRepository := transaction.NewDynamoDBTransactionRepository(dynamodbClient, dynamodbTableName, idempotencyStore)
	ledgerRepository := ledger.NewDynamoDBLedgerRepository(dynamodbClient, dynamodbTableName)

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
		CreateTransaction: ledger.NewCreateTransaction(ledgerRepository, accountRepository),
		GetTransaction:    ledger.NewGetTransaction(ledgerRepository),
		ListTransactions:  transaction.NewListTransactions(transactionRepository),
		UpdateTransaction: transaction.NewUpdateTransaction(transactionRepository, categoryRepository, legacyAccountRepository),
		DeleteTransaction: transaction.NewDeleteTransaction(transactionRepository, legacyAccountRepository),
	}

	router := NewRouter(identityProvider, *categoryHandler, *accountHandler, *transactionHandler)

	return &App{Handler: router}
}
