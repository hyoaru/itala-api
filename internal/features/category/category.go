package category

import (
	categoryrepositoryport "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	categoryusecases "github.com/hyoaru/itala-api/internal/features/category/application/usecases"
	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
	categoryrepositoryadapters "github.com/hyoaru/itala-api/internal/features/category/infrastructure/adapters/categoryrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type Category = entities.Category

type CategoryRepository = categoryrepositoryport.CategoryRepository

var ErrCategoryExists = categoryrepositoryport.ErrCategoryExists

func NewDynamoDBCategoryRepository(client dynamodbclient.DynamoDBClient, tableName string) CategoryRepository {
	r := categoryrepositoryadapters.NewDynamoDBCategoryRepository(client, tableName)
	return categoryrepositoryadapters.NewDecoratedCategoryRepository(r)
}

type (
	CreateCategoryRequest  = categoryusecases.CreateCategoryRequest
	CreateCategoryResponse = categoryusecases.CreateCategoryResponse
)

func NewCreateCategory(categoryRepository CategoryRepository) usecases.UseCase[CreateCategoryRequest, CreateCategoryResponse] {
	return categoryusecases.NewCreateCategory(categoryRepository)
}

type (
	ListCategoriesRequest  = categoryusecases.ListCategoriesRequest
	ListCategoriesResponse = categoryusecases.ListCategoriesResponse
)

func NewListCategories(categoryRepository CategoryRepository) usecases.UseCase[ListCategoriesRequest, categoryusecases.ListCategoriesResponse] {
	return categoryusecases.NewListCategories(categoryRepository)
}
