package identity

import "errors"

var (
	ErrTokenInvalid = errors.New("invalid token")
	ErrTokenExpired = errors.New("expired token")
)
