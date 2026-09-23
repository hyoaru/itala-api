package category

import (
	"context"
	"errors"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
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
		if got := repository.Create(context.Background(), "user-1", entity.Category{}); !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			transactWriteItems: func(
				ctx context.Context,
				input *dynamodbclient.TransactWriteItemsInput,
			) error {
				return boom
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		if got := repository.Create(context.Background(), "user-1", entity.Category{}); !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected transaction items", func(t *testing.T) {
		var captured *dynamodbclient.TransactWriteItemsInput

		client := &fakeDynamoDBClient{
			transactWriteItems: func(
				ctx context.Context,
				input *dynamodbclient.TransactWriteItemsInput,
			) error {
				captured = input
				return nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		category := entity.Category{
			ID:              "category-1",
			Name:            "Groceries",
			TransactionType: valueobject.TransactionTypeExpense,
			CreatedAt:       time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
			UpdatedAt:       time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		}

		_ = repository.Create(context.Background(), "user-1", category)

		if len(captured.TransactItems) != 2 {
			t.Fatalf("got %d transact items, want %d", len(captured.TransactItems), 2)
		}

		categoryPut := captured.TransactItems[0].Put
		if categoryPut == nil {
			t.Fatal("got nil category put, want non-nil")
		}

		if categoryPut.TableName != "table" {
			t.Errorf("got %v, want %v", categoryPut.TableName, "table")
		}

		if categoryPut.ConditionExpression != nil {
			t.Errorf("got %v, want %v", categoryPut.ConditionExpression, nil)
		}

		wantCategoryItem := map[string]any{
			"PK":               "USER#user-1",
			"SK":               "CATEGORY#category-1",
			"id":               "category-1",
			"name":             "Groceries",
			"transaction_type": "EXPENSE",
			"created_at":       category.CreatedAt.Format(time.RFC3339Nano),
			"updated_at":       category.UpdatedAt.Format(time.RFC3339Nano),
		}

		for key, want := range wantCategoryItem {
			if got := categoryPut.Item[key]; got != want {
				t.Errorf("category item %q: got %v, want %v", key, got, want)
			}
		}

		namePut := captured.TransactItems[1].Put
		if namePut == nil {
			t.Fatal("got nil name put, want non-nil")
		}

		if namePut.TableName != "table" {
			t.Errorf("got %v, want %v", namePut.TableName, "table")
		}

		wantNameItem := map[string]any{
			"PK":          "USER#user-1",
			"SK":          "CATEGORY_NAME#Groceries#EXPENSE",
			"category_id": "category-1",
		}

		for key, want := range wantNameItem {
			if got := namePut.Item[key]; got != want {
				t.Errorf("name item %q: got %v, want %v", key, got, want)
			}
		}

		if namePut.ConditionExpression == nil {
			t.Fatal("got nil condition expression, want non-nil")
		}

		if want := "attribute_not_exists(PK)"; *namePut.ConditionExpression != want {
			t.Errorf("got %v, want %v", *namePut.ConditionExpression, want)
		}
	})
}
