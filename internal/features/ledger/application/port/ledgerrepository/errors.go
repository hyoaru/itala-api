package ledgerrepository

import "errors"

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrIdempotencyMismatch = errors.New("idempotency payload mismatch")
	ErrConditionFailed     = errors.New("condition failed")
)
