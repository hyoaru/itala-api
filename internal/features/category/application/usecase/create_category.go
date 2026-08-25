package category

import (
	"context"
	"time"

	"github.com/google/uuid"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	categoryvo "github.com/hyoaru/itala-api/internal/features/category/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type CreateCategoryRequest struct {
	UserID string
	Name   string
	Type   valueobject.TransactionType
}

type CreateCategoryResponse entity.Category

type CreateCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewCreateCategory(categoryRepository categoryrepository.CategoryRepository) *CreateCategory {
	return &CreateCategory{categoryRepository: categoryRepository}
}

func (u *CreateCategory) Execute(ctx context.Context, request CreateCategoryRequest) (CreateCategoryResponse, error) {
	id := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()

	category := entity.Category{
		ID:              id.String(),
		Name:            request.Name,
		TransactionType: request.Type,
		Status:          categoryvo.StatusActive,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := u.categoryRepository.Create(ctx, request.UserID, category); err != nil {
		return CreateCategoryResponse{}, err
	}

	return CreateCategoryResponse(category), nil
}
