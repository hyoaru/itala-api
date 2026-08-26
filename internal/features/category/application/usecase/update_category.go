package category

import (
	"context"
	"time"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	categoryvalueobject "github.com/hyoaru/itala-api/internal/features/category/domain/valueobject"
)

type UpdateCategoryRequest struct {
	UserID string
	ID     string
	Name   string
	Status categoryvalueobject.Status
}

type UpdateCategoryResponse struct{}

type UpdateCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewUpdateCategory(categoryRepository categoryrepository.CategoryRepository) *UpdateCategory {
	return &UpdateCategory{categoryRepository: categoryRepository}
}

func (u *UpdateCategory) Execute(ctx context.Context, request UpdateCategoryRequest) (UpdateCategoryResponse, error) {
	now := time.Now().UTC()

	category := entity.Category{
		ID:        request.ID,
		Name:      request.Name,
		Status:    request.Status,
		UpdatedAt: now,
	}

	if err := u.categoryRepository.Update(ctx, request.UserID, category); err != nil {
		return UpdateCategoryResponse{}, err
	}

	return UpdateCategoryResponse{}, nil
}
