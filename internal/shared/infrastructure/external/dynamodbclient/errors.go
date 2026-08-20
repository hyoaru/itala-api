package dynamodbclient

import "errors"

var (
	ErrItemExists   = errors.New("item already exists")
	ErrItemNotFound = errors.New("item not found")
)
