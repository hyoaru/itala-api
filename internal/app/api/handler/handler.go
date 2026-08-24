package api

import (
	account "github.com/hyoaru/itala-api/internal/app/api/handler/account"
	category "github.com/hyoaru/itala-api/internal/app/api/handler/category"
	transaction "github.com/hyoaru/itala-api/internal/app/api/handler/transaction"
)

type (
	AccountHandler     = account.AccountHandler
	CategoryHandler    = category.CategoryHandler
	TransactionHandler = transaction.TransactionHandler
)
