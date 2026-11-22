package api

import (
	account "github.com/hyoaru/itala-api/internal/app/api/handler/account"
	transaction "github.com/hyoaru/itala-api/internal/app/api/handler/transaction"
	category "github.com/hyoaru/itala-api/internal/legacy/app/api/handler/category"
)

type (
	AccountHandler     = account.AccountHandler
	CategoryHandler    = category.CategoryHandler
	TransactionHandler = transaction.TransactionHandler
)
