package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
)

type GetCategoryRequest struct {
	UserID string
	ID     string
}

type GetCategoryResponse entity.Category

type GetCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewGetCategory(categoryRepository categoryrepository.CategoryRepository) *GetCategory {
	return &GetCategory{categoryRepository: categoryRepository}
}

func (u *GetCategory) Execute(ctx context.Context, request GetCategoryRequest) (GetCategoryResponse, error) {
	category, err := u.categoryRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return GetCategoryResponse{}, err
	}

	return GetCategoryResponse(category), nil
}
