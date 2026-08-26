package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
)

type RestoreCategoryRequest struct {
	UserID string
	ID     string
}

type RestoreCategoryResponse struct{}

type RestoreCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewRestoreCategory(categoryRepository categoryrepository.CategoryRepository) *RestoreCategory {
	return &RestoreCategory{categoryRepository: categoryRepository}
}

func (u *RestoreCategory) Execute(ctx context.Context, request RestoreCategoryRequest) (RestoreCategoryResponse, error) {
	if err := u.categoryRepository.Restore(ctx, request.UserID, request.ID); err != nil {
		return RestoreCategoryResponse{}, err
	}

	return RestoreCategoryResponse{}, nil
}
