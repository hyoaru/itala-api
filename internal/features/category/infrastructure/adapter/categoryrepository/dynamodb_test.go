package category

import (
	"context"
	"encoding/base64"
	"encoding/json"
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
	query              func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error)
	getItem            func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error
}

func (f *fakeDynamoDBClient) TransactWriteItems(
	ctx context.Context,
	input *dynamodbclient.TransactWriteItemsInput,
) error {
	return f.transactWriteItems(ctx, input)
}

func (f *fakeDynamoDBClient) Query(
	ctx context.Context,
	input *dynamodbclient.QueryInput,
	output any,
) (dynamodbclient.QueryOutput, error) {
	return f.query(ctx, input, output)
}

func (f *fakeDynamoDBClient) GetItem(
	ctx context.Context,
	input *dynamodbclient.GetItemInput,
	output any,
) error {
	return f.getItem(ctx, input, output)
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

func TestDynamoDBCategoryRepository_Find(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

		client := &fakeDynamoDBClient{
			query: func(
				ctx context.Context,
				input *dynamodbclient.QueryInput,
				output any,
			) (dynamodbclient.QueryOutput, error) {
				items := output.(*[]findCategoryItem)
				*items = append(*items, findCategoryItem{
					ID:              "category-1",
					Name:            "Groceries",
					TransactionType: "EXPENSE",
					CreatedAt:       createdAt.Format(time.RFC3339Nano),
					UpdatedAt:       createdAt.Format(time.RFC3339Nano),
				})
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		page, err := repository.Find(context.Background(), "user-1", port.CategoryQuery{})
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if len(page.Categories) != 1 {
			t.Fatalf("got %d categories, want %d", len(page.Categories), 1)
		}

		got := page.Categories[0]
		if got.ID != "category-1" {
			t.Errorf("got %v, want %v", got.ID, "category-1")
		}

		if got.Name != "Groceries" {
			t.Errorf("got %v, want %v", got.Name, "Groceries")
		}

		if got.TransactionType != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.TransactionType, valueobject.TransactionTypeExpense)
		}

		if !got.CreatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.CreatedAt, createdAt)
		}

		if !got.UpdatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, createdAt)
		}
	})

	t.Run("success with next cursor", func(t *testing.T) {
		lastEvaluatedKey := map[string]any{"PK": "USER#user-1", "SK": "CATEGORY#category-1"}

		client := &fakeDynamoDBClient{
			query: func(
				ctx context.Context,
				input *dynamodbclient.QueryInput,
				output any,
			) (dynamodbclient.QueryOutput, error) {
				return dynamodbclient.QueryOutput{LastEvaluatedKey: lastEvaluatedKey}, nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		page, err := repository.Find(context.Background(), "user-1", port.CategoryQuery{})
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if page.NextCursor == nil {
			t.Fatal("got nil next cursor, want non-nil")
		}

		decoded, err := base64.RawURLEncoding.DecodeString(*page.NextCursor)
		if err != nil {
			t.Fatalf("decode next cursor: %v", err)
		}

		var got map[string]any
		if err := json.Unmarshal(decoded, &got); err != nil {
			t.Fatalf("unmarshal next cursor: %v", err)
		}

		for key, want := range lastEvaluatedKey {
			if got[key] != want {
				t.Errorf("next cursor %q: got %v, want %v", key, got[key], want)
			}
		}
	})

	t.Run("query error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			query: func(
				ctx context.Context,
				input *dynamodbclient.QueryInput,
				output any,
			) (dynamodbclient.QueryOutput, error) {
				return dynamodbclient.QueryOutput{}, boom
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		_, err := repository.Find(context.Background(), "user-1", port.CategoryQuery{})
		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected query input", func(t *testing.T) {
		var captured *dynamodbclient.QueryInput

		client := &fakeDynamoDBClient{
			query: func(
				ctx context.Context,
				input *dynamodbclient.QueryInput,
				output any,
			) (dynamodbclient.QueryOutput, error) {
				captured = input
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		transactionType := valueobject.TransactionTypeExpense
		name := "Groceries"
		cursor := base64.RawURLEncoding.EncodeToString([]byte(`{"PK":"USER#user-1","SK":"CATEGORY#category-1"}`))

		query := port.CategoryQuery{
			Limit:           10,
			TransactionType: &transactionType,
			Name:            &name,
			Cursor:          &cursor,
		}

		_, _ = repository.Find(context.Background(), "user-1", query)

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if captured.Limit == nil || *captured.Limit != 10 {
			t.Errorf("got %v, want %v", captured.Limit, 10)
		}

		if captured.ScanIndexForward == nil || !*captured.ScanIndexForward {
			t.Errorf("got %v, want %v", captured.ScanIndexForward, true)
		}

		wantCondition := "PK = :pk AND begins_with(SK, :sk)"
		if captured.KeyConditionExpression == nil || *captured.KeyConditionExpression != wantCondition {
			t.Errorf("got %v, want %v", captured.KeyConditionExpression, wantCondition)
		}

		wantFilter := "(attribute_not_exists(deleted_at) OR attribute_type(deleted_at, :nullType)) AND transaction_type = :transaction_type AND #name = :name"
		if captured.FilterExpression == nil || *captured.FilterExpression != wantFilter {
			t.Errorf("got %v, want %v", captured.FilterExpression, wantFilter)
		}

		wantValues := map[string]any{
			":pk":               "USER#user-1",
			":sk":               "CATEGORY#",
			":nullType":         "NULL",
			":transaction_type": "EXPENSE",
			":name":             "Groceries",
		}

		for key, want := range wantValues {
			if got := captured.ExpressionAttributeValues[key]; got != want {
				t.Errorf("expression value %q: got %v, want %v", key, got, want)
			}
		}

		if len(captured.ExpressionAttributeNames) != 1 || captured.ExpressionAttributeNames["#name"] != "name" {
			t.Errorf("got %v, want %v", captured.ExpressionAttributeNames, map[string]string{"#name": "name"})
		}

		if got := captured.ExclusiveStartKey["PK"]; got != "USER#user-1" {
			t.Errorf("start key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.ExclusiveStartKey["SK"]; got != "CATEGORY#category-1" {
			t.Errorf("start key %q: got %v, want %v", "SK", got, "CATEGORY#category-1")
		}
	})
}

func TestFindCategoryItem_toDomain(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success", func(t *testing.T) {
		item := findCategoryItem{
			ID:              "category-1",
			Name:            "Groceries",
			TransactionType: "EXPENSE",
			CreatedAt:       timestamp,
			UpdatedAt:       timestamp,
		}

		got, err := item.toDomain()
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "category-1" {
			t.Errorf("got %v, want %v", got.ID, "category-1")
		}

		if got.Name != "Groceries" {
			t.Errorf("got %v, want %v", got.Name, "Groceries")
		}

		if got.TransactionType != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.TransactionType, valueobject.TransactionTypeExpense)
		}

		if !got.CreatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.CreatedAt, createdAt)
		}

		if !got.UpdatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, createdAt)
		}

		if got.DeletedAt != nil {
			t.Errorf("got %v, want %v", got.DeletedAt, nil)
		}
	})

	t.Run("deleted at", func(t *testing.T) {
		item := findCategoryItem{
			ID:              "category-1",
			Name:            "Groceries",
			TransactionType: "EXPENSE",
			DeletedAt:       &timestamp,
			CreatedAt:       timestamp,
			UpdatedAt:       timestamp,
		}

		got, err := item.toDomain()
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.DeletedAt == nil {
			t.Fatal("got nil deleted at, want non-nil")
		}

		if !got.DeletedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.DeletedAt, createdAt)
		}
	})

	t.Run("invalid created_at", func(t *testing.T) {
		item := findCategoryItem{
			CreatedAt: "nope",
			UpdatedAt: timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid updated_at", func(t *testing.T) {
		item := findCategoryItem{
			CreatedAt: timestamp,
			UpdatedAt: "nope",
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid deleted_at", func(t *testing.T) {
		invalid := "nope"
		item := findCategoryItem{
			DeletedAt: &invalid,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})
}

func TestDynamoDBCategoryRepository_FindOne(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

		client := &fakeDynamoDBClient{
			getItem: func(
				ctx context.Context,
				input *dynamodbclient.GetItemInput,
				output any,
			) error {
				item := output.(*findCategoryItem)
				*item = findCategoryItem{
					ID:              "category-1",
					Name:            "Groceries",
					TransactionType: "EXPENSE",
					CreatedAt:       createdAt.Format(time.RFC3339Nano),
					UpdatedAt:       createdAt.Format(time.RFC3339Nano),
				}
				return nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		got, err := repository.FindOne(context.Background(), "user-1", "category-1")
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "category-1" {
			t.Errorf("got %v, want %v", got.ID, "category-1")
		}

		if got.Name != "Groceries" {
			t.Errorf("got %v, want %v", got.Name, "Groceries")
		}

		if got.TransactionType != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.TransactionType, valueobject.TransactionTypeExpense)
		}

		if !got.CreatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.CreatedAt, createdAt)
		}

		if !got.UpdatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, createdAt)
		}
	})

	t.Run("category not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(
				ctx context.Context,
				input *dynamodbclient.GetItemInput,
				output any,
			) error {
				return dynamodbclient.ErrItemNotFound
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		want := port.ErrCategoryNotFound
		_, got := repository.FindOne(context.Background(), "user-1", "category-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("category deleted", func(t *testing.T) {
		deletedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
		deletedAtStr := deletedAt.Format(time.RFC3339Nano)

		client := &fakeDynamoDBClient{
			getItem: func(
				ctx context.Context,
				input *dynamodbclient.GetItemInput,
				output any,
			) error {
				item := output.(*findCategoryItem)
				*item = findCategoryItem{
					ID:              "category-1",
					Name:            "Groceries",
					TransactionType: "EXPENSE",
					DeletedAt:       &deletedAtStr,
					CreatedAt:       deletedAtStr,
					UpdatedAt:       deletedAtStr,
				}
				return nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		want := port.ErrCategoryNotFound
		_, got := repository.FindOne(context.Background(), "user-1", "category-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			getItem: func(
				ctx context.Context,
				input *dynamodbclient.GetItemInput,
				output any,
			) error {
				return boom
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		_, got := repository.FindOne(context.Background(), "user-1", "category-1")

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected get item input", func(t *testing.T) {
		var captured *dynamodbclient.GetItemInput

		client := &fakeDynamoDBClient{
			getItem: func(
				ctx context.Context,
				input *dynamodbclient.GetItemInput,
				output any,
			) error {
				captured = input
				return nil
			},
		}

		repository := NewDynamoDBCategoryRepository(client, "table")

		_, _ = repository.FindOne(context.Background(), "user-1", "category-1")

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.Key["SK"]; got != "CATEGORY#category-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "CATEGORY#category-1")
		}
	})
}
