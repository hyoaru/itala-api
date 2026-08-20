package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CreateCategoryRequest struct {
	UserID          string
	Name            string
	TransactionType valueobjects.TransactionType
}

type CreateCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewCreateCategory(categoryRepository categoryrepository.CategoryRepository) *CreateCategory {
	return &CreateCategory{categoryRepository: categoryRepository}
}

func (u *CreateCategory) Execute(ctx context.Context, request CreateCategoryRequest) (struct{}, error) {
	err := u.categoryRepository.Create(
		ctx,
		request.UserID,
		request.Name,
		request.TransactionType,
	)
	if err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
