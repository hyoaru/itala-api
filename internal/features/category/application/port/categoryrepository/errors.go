package category

import "errors"

var (
	ErrCategoryExists         = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryArchived       = errors.New("category is archived")
	ErrCategoryNotActive      = errors.New("category is not active")
	ErrCategoryNotArchived    = errors.New("category is not archived")
	ErrConcurrentModification = errors.New("concurrent modification")
)
