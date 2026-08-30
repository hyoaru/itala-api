package account

import "errors"

var (
	ErrAccountExists          = errors.New("account already exists")
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountArchived        = errors.New("account is archived")
	ErrConcurrentModification = errors.New("concurrent modification")
)
