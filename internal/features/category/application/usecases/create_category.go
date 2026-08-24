package category

import (
	"context"
	"time"

	"github.com/google/uuid"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CreateCategoryRequest struct {
	UserID string
	Name   string
	Type   valueobjects.TransactionType
}

type CreateCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewCreateCategory(categoryRepository categoryrepository.CategoryRepository) *CreateCategory {
	return &CreateCategory{categoryRepository: categoryRepository}
}

func (u *CreateCategory) Execute(ctx context.Context, request CreateCategoryRequest) (struct{}, error) {
	id := uuid.New()
	now := time.Now().UTC()

	category := entities.Category{
		ID:              id.String(),
		Name:            request.Name,
		TransactionType: request.Type,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := u.categoryRepository.Create(ctx, request.UserID, category); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
