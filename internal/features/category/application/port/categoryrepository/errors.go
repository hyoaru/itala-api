package category

import "errors"

var (
	ErrCategoryExists         = errors.New("category already exists")
	ErrCategoryNotFound       = errors.New("category not found")
	ErrCategoryDeleted         = errors.New("category is deleted")
	ErrConcurrentModification = errors.New("concurrent modification")
)
