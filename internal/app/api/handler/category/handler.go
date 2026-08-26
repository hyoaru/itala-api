package api

import (
	"github.com/hyoaru/itala-api/internal/features/category"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
)

type CategoryHandler struct {
	CreateCategory  usecase.UseCase[category.CreateCategoryRequest, category.CreateCategoryResponse]
	ListCategories  usecase.UseCase[category.ListCategoriesRequest, category.ListCategoriesResponse]
	GetCategory     usecase.UseCase[category.GetCategoryRequest, category.GetCategoryResponse]
	UpdateCategory  usecase.UseCase[category.UpdateCategoryRequest, category.UpdateCategoryResponse]
	ArchiveCategory usecase.UseCase[category.ArchiveCategoryRequest, category.ArchiveCategoryResponse]
	RestoreCategory usecase.UseCase[category.RestoreCategoryRequest, category.RestoreCategoryResponse]
}
