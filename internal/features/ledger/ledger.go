package ledger

import (
	account "github.com/hyoaru/itala-api/internal/features/account"
	ledgerrepositoryport "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	ledgerusecase "github.com/hyoaru/itala-api/internal/features/ledger/application/usecase"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
	ledgerrepositoryadapter "github.com/hyoaru/itala-api/internal/features/ledger/infrastructure/adapter/ledgerrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type Transaction = entity.Transaction
type Posting = entity.Posting
type BalanceDelta = entity.BalanceDelta
type TransactionSummary = entity.TransactionSummary
type AccountStatementEntry = entity.AccountStatementEntry
type IdempotencyRecord = ledgerrepositoryport.IdempotencyRecord

type LedgerRepository = ledgerrepositoryport.LedgerRepository
type RecordTransactionInput = ledgerrepositoryport.RecordTransactionInput

var (
	ErrTransactionNotFound    = ledgerrepositoryport.ErrTransactionNotFound
	ErrIdempotencyMismatch    = ledgerrepositoryport.ErrIdempotencyMismatch
	ErrConditionFailed        = ledgerrepositoryport.ErrConditionFailed
	ErrInvalidTransaction     = ledgerusecase.ErrInvalidTransaction
	ErrInvalidCategoryAccount = ledgerusecase.ErrInvalidCategoryAccount
)

func NewDynamoDBLedgerRepository(client dynamodbclient.DynamoDBClient, tableName string) LedgerRepository {
	return ledgerrepositoryadapter.NewDecoratedLedgerRepository(ledgerrepositoryadapter.NewDynamoDBLedgerRepository(client, tableName))
}

type (
	CreateTransactionRequest  = ledgerusecase.CreateTransactionRequest
	CreateTransactionResponse = ledgerusecase.CreateTransactionResponse
)

func NewCreateTransaction(
	ledgerRepository LedgerRepository,
	accountRepository account.AccountRepository,
) usecase.UseCase[CreateTransactionRequest, CreateTransactionResponse] {
	return ledgerusecase.NewCreateTransaction(ledgerRepository, accountRepository)
}

type (
	GetTransactionRequest  = ledgerusecase.GetTransactionRequest
	GetTransactionResponse = ledgerusecase.GetTransactionResponse
)

func NewGetTransaction(ledgerRepository LedgerRepository) usecase.UseCase[GetTransactionRequest, GetTransactionResponse] {
	return ledgerusecase.NewGetTransaction(ledgerRepository)
}
