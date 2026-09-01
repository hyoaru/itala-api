package category

import (
	"context"
	"time"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
)

type UpdateCategoryRequest struct {
	UserID string
	ID     string
	Name   string
}

type UpdateCategoryResponse struct{}

type UpdateCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewUpdateCategory(categoryRepository categoryrepository.CategoryRepository) *UpdateCategory {
	return &UpdateCategory{categoryRepository: categoryRepository}
}

func (u *UpdateCategory) Execute(ctx context.Context, request UpdateCategoryRequest) (UpdateCategoryResponse, error) {
	current, err := u.categoryRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return UpdateCategoryResponse{}, err
	}

	if current.DeletedAt != nil {
		return UpdateCategoryResponse{}, categoryrepository.ErrCategoryNotFound
	}

	now := time.Now().UTC()

	category := entity.Category{
		ID:        request.ID,
		Name:      request.Name,
		UpdatedAt: now,
	}

	if err := u.categoryRepository.Update(ctx, request.UserID, category); err != nil {
		return UpdateCategoryResponse{}, err
	}

	return UpdateCategoryResponse{}, nil
}
