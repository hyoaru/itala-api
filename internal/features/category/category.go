package category

import (
	categoryrepositoryport "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	categoryusecase "github.com/hyoaru/itala-api/internal/features/category/application/usecase"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/features/category/domain/valueobject"
	categoryrepositoryadapter "github.com/hyoaru/itala-api/internal/features/category/infrastructure/adapter/categoryrepository"
	"github.com/hyoaru/itala-api/internal/shared/application/usecase"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type (
	Category = entity.Category
	Status   = valueobject.Status
)

const (
	StatusActive   = valueobject.StatusActive
	StatusArchived = valueobject.StatusArchived
)

type CategoryRepository = categoryrepositoryport.CategoryRepository

var (
	ErrCategoryExists   = categoryrepositoryport.ErrCategoryExists
	ErrCategoryNotFound = categoryrepositoryport.ErrCategoryNotFound
	ErrCategoryArchived = categoryrepositoryport.ErrCategoryArchived
)

func NewDynamoDBCategoryRepository(client dynamodbclient.DynamoDBClient, tableName string) CategoryRepository {
	r := categoryrepositoryadapter.NewDynamoDBCategoryRepository(client, tableName)
	return categoryrepositoryadapter.NewDecoratedCategoryRepository(r)
}

type (
	CreateCategoryRequest  = categoryusecase.CreateCategoryRequest
	CreateCategoryResponse = categoryusecase.CreateCategoryResponse
)

func NewCreateCategory(categoryRepository CategoryRepository) usecase.UseCase[CreateCategoryRequest, CreateCategoryResponse] {
	return categoryusecase.NewCreateCategory(categoryRepository)
}

type (
	ListCategoriesRequest  = categoryusecase.ListCategoriesRequest
	ListCategoriesResponse = categoryusecase.ListCategoriesResponse
)

func NewListCategories(categoryRepository CategoryRepository) usecase.UseCase[ListCategoriesRequest, categoryusecase.ListCategoriesResponse] {
	return categoryusecase.NewListCategories(categoryRepository)
}

type (
	UpdateCategoryRequest  = categoryusecase.UpdateCategoryRequest
	UpdateCategoryResponse = categoryusecase.UpdateCategoryResponse
)

func NewUpdateCategory(categoryRepository CategoryRepository) usecase.UseCase[UpdateCategoryRequest, UpdateCategoryResponse] {
	return categoryusecase.NewUpdateCategory(categoryRepository)
}

type (
	GetCategoryRequest  = categoryusecase.GetCategoryRequest
	GetCategoryResponse = categoryusecase.GetCategoryResponse
)

func NewGetCategory(categoryRepository CategoryRepository) usecase.UseCase[GetCategoryRequest, GetCategoryResponse] {
	return categoryusecase.NewGetCategory(categoryRepository)
}
