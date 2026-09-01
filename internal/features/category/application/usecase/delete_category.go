package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
)

type DeleteCategoryRequest struct {
	UserID string
	ID     string
}

type DeleteCategoryResponse struct{}

type DeleteCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewDeleteCategory(categoryRepository categoryrepository.CategoryRepository) *DeleteCategory {
	return &DeleteCategory{categoryRepository: categoryRepository}
}

func (u *DeleteCategory) Execute(ctx context.Context, request DeleteCategoryRequest) (DeleteCategoryResponse, error) {
	current, err := u.categoryRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return DeleteCategoryResponse{}, err
	}

	if current.DeletedAt != nil {
		return DeleteCategoryResponse{}, categoryrepository.ErrCategoryNotFound
	}

	if err := u.categoryRepository.Delete(ctx, request.UserID, request.ID); err != nil {
		return DeleteCategoryResponse{}, err
	}

	return DeleteCategoryResponse{}, nil
}
