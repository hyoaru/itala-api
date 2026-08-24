package category

import "errors"

var (
	ErrCategoryExists   = errors.New("category already exists")
	ErrCategoryNotFound = errors.New("category not found")
)
