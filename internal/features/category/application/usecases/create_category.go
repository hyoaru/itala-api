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
	request            CreateCategoryRequest
	categoryRepository categoryrepository.CategoryRepository
}

func NewCreateCategory(
	request CreateCategoryRequest,
	categoryRepository categoryrepository.CategoryRepository,
) *CreateCategory {
	return &CreateCategory{
		request:            request,
		categoryRepository: categoryRepository,
	}
}

func (u *CreateCategory) Execute(ctx context.Context) error {
	return u.categoryRepository.Create(
		ctx,
		u.request.UserID,
		u.request.Name,
		u.request.TransactionType,
	)
}
