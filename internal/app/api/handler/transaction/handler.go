package api

import (
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
)

type TransactionHandler struct {
	CreateTransaction usecase.UseCase[transaction.CreateTransactionRequest, transaction.CreateTransactionResponse]
	ListTransactions  usecase.UseCase[transaction.ListTransactionsRequest, transaction.ListTransactionsResponse]
	UpdateTransaction usecase.UseCase[transaction.UpdateTransactionRequest, transaction.UpdateTransactionResponse]
}
