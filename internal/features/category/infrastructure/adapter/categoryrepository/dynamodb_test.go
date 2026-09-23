package category

import (
	"context"
	"errors"
	"testing"

	port "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type fakeDynamoDBClient struct {
	dynamodbclient.DynamoDBClient
	transactWriteItems func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error
}

func (f *fakeDynamoDBClient) TransactWriteItems(
	ctx context.Context,
	input *dynamodbclient.TransactWriteItemsInput,
) error {
	return f.transactWriteItems(ctx, input)
}

func TestDynamoDBCategoryRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			transactWriteItems: func(
				ctx context.Context,
				input *dynamodbclient.TransactWriteItemsInput,
			) error {
				return nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		if got := repository.Create(context.Background(), "user-1", entity.Category{}); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("category already exists", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			transactWriteItems: func(
				ctx context.Context,
				input *dynamodbclient.TransactWriteItemsInput,
			) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		want := port.ErrCategoryExists
		got := repository.Create(context.Background(), "user-1", entity.Category{})

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
}
