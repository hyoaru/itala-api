package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
)

type ArchiveCategoryRequest struct {
	UserID string
	ID     string
}

type ArchiveCategoryResponse struct{}

type ArchiveCategory struct {
	categoryRepository categoryrepository.CategoryRepository
}

func NewArchiveCategory(categoryRepository categoryrepository.CategoryRepository) *ArchiveCategory {
	return &ArchiveCategory{categoryRepository: categoryRepository}
}

func (u *ArchiveCategory) Execute(ctx context.Context, request ArchiveCategoryRequest) (ArchiveCategoryResponse, error) {
	if err := u.categoryRepository.Archive(ctx, request.UserID, request.ID); err != nil {
		return ArchiveCategoryResponse{}, err
	}

	return ArchiveCategoryResponse{}, nil
}
