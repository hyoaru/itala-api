package account

import "errors"

var (
	ErrAccountExists          = errors.New("account already exists")
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountDeleted          = errors.New("account is deleted")
	ErrConcurrentModification = errors.New("concurrent modification")
)
