package transaction

import "errors"

var (
	ErrTransactionExists   = errors.New("transaction already exists")
	ErrTransactionNotFound = errors.New("transaction not found")
)
