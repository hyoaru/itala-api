package api

import (
	"github.com/hyoaru/itala-api/internal/features/ledger"
	legacytransaction "github.com/hyoaru/itala-api/internal/legacy/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
)

type TransactionHandler struct {
	CreateTransaction usecase.UseCase[ledger.CreateTransactionRequest, ledger.CreateTransactionResponse]
	GetTransaction    usecase.UseCase[ledger.GetTransactionRequest, ledger.GetTransactionResponse]

	// shortcut: list, update, and delete stay on legacy use cases until their tickets land.
	ListTransactions  usecase.UseCase[legacytransaction.ListTransactionsRequest, legacytransaction.ListTransactionsResponse]
	UpdateTransaction usecase.UseCase[legacytransaction.UpdateTransactionRequest, legacytransaction.UpdateTransactionResponse]
	DeleteTransaction usecase.UseCase[legacytransaction.DeleteTransactionRequest, legacytransaction.DeleteTransactionResponse]
}
