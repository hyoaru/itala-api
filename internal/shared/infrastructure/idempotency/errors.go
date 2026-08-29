package idempotency

import "errors"

var (
	ErrItemNotFound     = errors.New("idempotency item not found")
	ErrInvalidLockToken = errors.New("invalid lock token")
)
