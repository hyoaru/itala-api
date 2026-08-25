package api

import (
	"github.com/hyoaru/itala-api/internal/features/category"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
)

type CategoryHandler struct {
	CreateCategory usecases.UseCase[category.CreateCategoryRequest, category.CreateCategoryResponse]
	ListCategories usecases.UseCase[category.ListCategoriesRequest, category.ListCategoriesResponse]
}
