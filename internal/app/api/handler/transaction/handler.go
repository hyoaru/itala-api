package api

import (
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
)

type TransactionHandler struct {
	CreateTransaction usecases.UseCase[transaction.CreateTransactionRequest, struct{}]
	ListTransactions  usecases.UseCase[transaction.ListTransactionsRequest, transaction.ListTransactionsResponse]
}
