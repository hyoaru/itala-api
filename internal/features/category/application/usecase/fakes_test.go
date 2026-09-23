package category

import (
	"context"

	categoryrepository "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
)

type fakeCategoryRepository struct {
	categoryrepository.CategoryRepository
	create  func(ctx context.Context, userID string, category entity.Category) error
	find    func(ctx context.Context, userID string, query categoryrepository.CategoryQuery) (categoryrepository.CategoryPage, error)
	findOne func(ctx context.Context, userID string, categoryID string) (entity.Category, error)
	update  func(ctx context.Context, userID string, category entity.Category) error
	delete  func(ctx context.Context, userID string, categoryID string) error
}

func (f *fakeCategoryRepository) Create(ctx context.Context, userID string, category entity.Category) error {
	return f.create(ctx, userID, category)
}

func (f *fakeCategoryRepository) Find(
	ctx context.Context,
	userID string,
	query categoryrepository.CategoryQuery,
) (categoryrepository.CategoryPage, error) {
	return f.find(ctx, userID, query)
}

func (f *fakeCategoryRepository) FindOne(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
	return f.findOne(ctx, userID, categoryID)
}

func (f *fakeCategoryRepository) Update(ctx context.Context, userID string, category entity.Category) error {
	return f.update(ctx, userID, category)
}

func (f *fakeCategoryRepository) Delete(ctx context.Context, userID string, categoryID string) error {
	return f.delete(ctx, userID, categoryID)
}
