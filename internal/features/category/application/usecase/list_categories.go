package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type ListCategoriesRequest struct {
	UserID          string
	Limit           int32
	Name            *string
	TransactionType *valueobject.TransactionType
	Cursor          *string
}

type ListCategoriesResponse struct {
	Categories []entity.Category
	NextCursor *string
}

type ListCategories struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewListCategories(categoryRepository categoryrepository.CategoryRepository) *ListCategories {
	return &ListCategories{categoryRepository: categoryRepository}
}

func (u *ListCategories) Execute(ctx context.Context, request ListCategoriesRequest) (ListCategoriesResponse, error) {
	query := categoryrepository.CategoryQuery{
		Limit:           request.Limit,
		Name:            request.Name,
		TransactionType: request.TransactionType,
		Cursor:          request.Cursor,
	}

	page, err := u.categoryRepository.Find(ctx, request.UserID, query)
	if err != nil {
		return ListCategoriesResponse{}, err
	}

	return ListCategoriesResponse{Categories: page.Categories, NextCursor: page.NextCursor}, nil
}
