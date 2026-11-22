package api

import (
	account "github.com/hyoaru/itala-api/internal/legacy/app/api/handler/account"
	category "github.com/hyoaru/itala-api/internal/legacy/app/api/handler/category"
	transaction "github.com/hyoaru/itala-api/internal/legacy/app/api/handler/transaction"
)

type (
	AccountHandler     = account.AccountHandler
	CategoryHandler    = category.CategoryHandler
	TransactionHandler = transaction.TransactionHandler
)
