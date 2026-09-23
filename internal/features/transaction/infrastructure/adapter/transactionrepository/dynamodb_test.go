package transaction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	port "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type fakeDynamoDBClient struct {
	dynamodbclient.DynamoDBClient
	putItem    func(ctx context.Context, input *dynamodbclient.PutItemInput) error
	query      func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error)
	getItem    func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error
	updateItem func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error
}

func (f *fakeDynamoDBClient) PutItem(ctx context.Context, input *dynamodbclient.PutItemInput) error {
	return f.putItem(ctx, input)
}

func (f *fakeDynamoDBClient) Query(
	ctx context.Context,
	input *dynamodbclient.QueryInput,
	output any,
) (dynamodbclient.QueryOutput, error) {
	return f.query(ctx, input, output)
}

func (f *fakeDynamoDBClient) GetItem(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
	return f.getItem(ctx, input, output)
}

func (f *fakeDynamoDBClient) UpdateItem(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
	return f.updateItem(ctx, input)
}

func TestDynamoDBTransactionRepository_Create(t *testing.T) {
	amount, err := valueobject.NewDecimal("10")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		if got := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1"); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("transaction already exists", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return dynamodbclient.ErrItemExists
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		want := port.ErrTransactionExists
		got := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return boom
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		got := repository.Create(context.Background(), "user-1", entity.Transaction{}, "idem-1")

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected item", func(t *testing.T) {
		var captured *dynamodbclient.PutItemInput

		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				captured = input
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		transaction := entity.Transaction{
			ID:          "transaction-1",
			Amount:      amount,
			Type:        valueobject.TransactionTypeExpense,
			AccountID:   "account-1",
			CategoryID:  "category-1",
			Description: "Lunch",
			OccurredAt:  createdAt,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}

		_ = repository.Create(context.Background(), "user-1", transaction, "idem-1")

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		gsiSortKey := "TRANSACTION#" + createdAt.Format(time.RFC3339Nano) + "transaction-1"

		wantItem := map[string]any{
			"PK":          "USER#user-1",
			"SK":          "TRANSACTION#transaction-1",
			"GSI1PK":      "USER#user-1",
			"GSI1SK":      gsiSortKey,
			"GSI2PK":      "USER#user-1#TYPE#EXPENSE",
			"GSI2SK":      gsiSortKey,
			"GSI3PK":      "USER#user-1#ACCOUNT#account-1",
			"GSI3SK":      gsiSortKey,
			"GSI4PK":      "USER#user-1#CATEGORY#category-1",
			"GSI4SK":      gsiSortKey,
			"id":          "transaction-1",
			"amount":      dynamodbclient.Decimal(amount),
			"type":        "EXPENSE",
			"account_id":  "account-1",
			"category_id": "category-1",
			"description": "Lunch",
			"occurred_at": createdAt.Format(time.RFC3339Nano),
			"created_at":  createdAt.Format(time.RFC3339Nano),
			"updated_at":  createdAt.Format(time.RFC3339Nano),
		}

		for key, want := range wantItem {
			if got := captured.Item[key]; got != want {
				t.Errorf("item %q: got %v, want %v", key, got, want)
			}
		}
	})
}

func TestDynamoDBTransactionRepository_Find(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				items := output.(*[]findTransactionItem)
				*items = append(*items, findTransactionItem{
					ID:          "transaction-1",
					Amount:      attributevalue.Number("10"),
					Type:        "EXPENSE",
					AccountID:   "account-1",
					CategoryID:  "category-1",
					Description: "Lunch",
					OccurredAt:  timestamp,
					CreatedAt:   timestamp,
					UpdatedAt:   timestamp,
				})
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		page, err := repository.Find(context.Background(), "user-1", port.TransactionQuery{})
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if len(page.Transactions) != 1 {
			t.Fatalf("got %d transactions, want %d", len(page.Transactions), 1)
		}

		got := page.Transactions[0]
		if got.ID != "transaction-1" {
			t.Errorf("got %v, want %v", got.ID, "transaction-1")
		}

		if got.Amount.String() != "10" {
			t.Errorf("got %v, want %v", got.Amount.String(), "10")
		}

		if got.Type != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.Type, valueobject.TransactionTypeExpense)
		}

		if got.AccountID != "account-1" || got.CategoryID != "category-1" {
			t.Errorf("got %v/%v, want account-1/category-1", got.AccountID, got.CategoryID)
		}
	})

	t.Run("success with next cursor", func(t *testing.T) {
		lastEvaluatedKey := map[string]any{"PK": "USER#user-1", "SK": "TRANSACTION#transaction-1"}

		client := &fakeDynamoDBClient{
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				return dynamodbclient.QueryOutput{LastEvaluatedKey: lastEvaluatedKey}, nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		page, err := repository.Find(context.Background(), "user-1", port.TransactionQuery{})
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
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				return dynamodbclient.QueryOutput{}, boom
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		_, err := repository.Find(context.Background(), "user-1", port.TransactionQuery{})
		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected query by category", func(t *testing.T) {
		var captured *dynamodbclient.QueryInput

		client := &fakeDynamoDBClient{
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				captured = input
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		transactionType := valueobject.TransactionTypeExpense
		categoryID := "category-1"
		accountID := "account-1"
		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		cursor := base64.RawURLEncoding.EncodeToString([]byte(`{"PK":"USER#user-1"}`))

		query := port.TransactionQuery{
			Limit:      10,
			Type:       &transactionType,
			CategoryID: &categoryID,
			AccountID:  &accountID,
			From:       &from,
			To:         &to,
			Cursor:     &cursor,
		}

		_, _ = repository.Find(context.Background(), "user-1", query)

		if captured.IndexName == nil || *captured.IndexName != "TransactionByCategory" {
			t.Errorf("got %v, want %v", captured.IndexName, "TransactionByCategory")
		}

		if captured.ScanIndexForward == nil || *captured.ScanIndexForward {
			t.Errorf("got %v, want %v", captured.ScanIndexForward, false)
		}

		wantCondition := "GSI4PK = :pk AND GSI4SK BETWEEN :from AND :to"
		if captured.KeyConditionExpression == nil || *captured.KeyConditionExpression != wantCondition {
			t.Errorf("got %v, want %v", captured.KeyConditionExpression, wantCondition)
		}

		wantFilter := "(attribute_not_exists(deleted_at) OR attribute_type(deleted_at, :nullType)) AND #type = :type AND category_id = :category_id AND account_id = :account_id"
		if captured.FilterExpression == nil || *captured.FilterExpression != wantFilter {
			t.Errorf("got %v, want %v", captured.FilterExpression, wantFilter)
		}

		wantValues := map[string]any{
			":pk":          "USER#user-1#CATEGORY#category-1",
			":from":        "TRANSACTION#" + from.Format(time.RFC3339Nano),
			":to":          "TRANSACTION#" + to.Format(time.RFC3339Nano) + "~",
			":nullType":    "NULL",
			":type":        "EXPENSE",
			":category_id": "category-1",
			":account_id":  "account-1",
		}

		for key, want := range wantValues {
			if got := captured.ExpressionAttributeValues[key]; got != want {
				t.Errorf("expression value %q: got %v, want %v", key, got, want)
			}
		}

		if len(captured.ExpressionAttributeNames) != 1 || captured.ExpressionAttributeNames["#type"] != "type" {
			t.Errorf("got %v, want %v", captured.ExpressionAttributeNames, map[string]string{"#type": "type"})
		}
	})

	t.Run("writes expected query by account", func(t *testing.T) {
		var captured *dynamodbclient.QueryInput

		client := &fakeDynamoDBClient{
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				captured = input
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		accountID := "account-1"
		query := port.TransactionQuery{Limit: 10, AccountID: &accountID}

		_, _ = repository.Find(context.Background(), "user-1", query)

		if captured.IndexName == nil || *captured.IndexName != "TransactionByAccount" {
			t.Errorf("got %v, want %v", captured.IndexName, "TransactionByAccount")
		}

		wantCondition := "GSI3PK = :pk AND begins_with(GSI3SK, :sk)"
		if captured.KeyConditionExpression == nil || *captured.KeyConditionExpression != wantCondition {
			t.Errorf("got %v, want %v", captured.KeyConditionExpression, wantCondition)
		}

		if got := captured.ExpressionAttributeValues[":pk"]; got != "USER#user-1#ACCOUNT#account-1" {
			t.Errorf("pk: got %v, want %v", got, "USER#user-1#ACCOUNT#account-1")
		}

		if got := captured.ExpressionAttributeValues[":sk"]; got != "TRANSACTION#" {
			t.Errorf("sk: got %v, want %v", got, "TRANSACTION#")
		}
	})

	t.Run("writes expected query by type", func(t *testing.T) {
		var captured *dynamodbclient.QueryInput

		client := &fakeDynamoDBClient{
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				captured = input
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		transactionType := valueobject.TransactionTypeIncome
		query := port.TransactionQuery{Limit: 10, Type: &transactionType}

		_, _ = repository.Find(context.Background(), "user-1", query)

		if captured.IndexName == nil || *captured.IndexName != "TransactionByType" {
			t.Errorf("got %v, want %v", captured.IndexName, "TransactionByType")
		}

		wantCondition := "GSI2PK = :pk AND begins_with(GSI2SK, :sk)"
		if captured.KeyConditionExpression == nil || *captured.KeyConditionExpression != wantCondition {
			t.Errorf("got %v, want %v", captured.KeyConditionExpression, wantCondition)
		}

		if got := captured.ExpressionAttributeValues[":pk"]; got != "USER#user-1#TYPE#INCOME" {
			t.Errorf("pk: got %v, want %v", got, "USER#user-1#TYPE#INCOME")
		}
	})

	t.Run("writes expected query default", func(t *testing.T) {
		var captured *dynamodbclient.QueryInput

		client := &fakeDynamoDBClient{
			query: func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
				captured = input
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
		query := port.TransactionQuery{Limit: 10, From: &from, To: &to}

		_, _ = repository.Find(context.Background(), "user-1", query)

		if captured.IndexName == nil || *captured.IndexName != "TransactionByOccurredAt" {
			t.Errorf("got %v, want %v", captured.IndexName, "TransactionByOccurredAt")
		}

		wantCondition := "GSI1PK = :pk AND GSI1SK BETWEEN :from AND :to"
		if captured.KeyConditionExpression == nil || *captured.KeyConditionExpression != wantCondition {
			t.Errorf("got %v, want %v", captured.KeyConditionExpression, wantCondition)
		}

		if got := captured.ExpressionAttributeValues[":pk"]; got != "USER#user-1" {
			t.Errorf("pk: got %v, want %v", got, "USER#user-1")
		}
	})
}

func TestDynamoDBTransactionRepository_FindOne(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findTransactionItem)
				*item = findTransactionItem{
					ID:         "transaction-1",
					Amount:     attributevalue.Number("10"),
					Type:       "EXPENSE",
					AccountID:  "account-1",
					CategoryID: "category-1",
					OccurredAt: timestamp,
					CreatedAt:  timestamp,
					UpdatedAt:  timestamp,
				}
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		got, err := repository.FindOne(context.Background(), "user-1", "transaction-1")
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "transaction-1" {
			t.Errorf("got %v, want %v", got.ID, "transaction-1")
		}

		if got.Amount.String() != "10" {
			t.Errorf("got %v, want %v", got.Amount.String(), "10")
		}

		if got.Type != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.Type, valueobject.TransactionTypeExpense)
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return dynamodbclient.ErrItemNotFound
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		want := port.ErrTransactionNotFound
		_, got := repository.FindOne(context.Background(), "user-1", "transaction-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("transaction deleted", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findTransactionItem)
				*item = findTransactionItem{
					ID:         "transaction-1",
					Amount:     attributevalue.Number("10"),
					Type:       "EXPENSE",
					OccurredAt: timestamp,
					CreatedAt:  timestamp,
					UpdatedAt:  timestamp,
					DeletedAt:  &timestamp,
				}
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		want := port.ErrTransactionNotFound
		_, got := repository.FindOne(context.Background(), "user-1", "transaction-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return boom
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		_, got := repository.FindOne(context.Background(), "user-1", "transaction-1")

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected get item input", func(t *testing.T) {
		var captured *dynamodbclient.GetItemInput

		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				captured = input
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		_, _ = repository.FindOne(context.Background(), "user-1", "transaction-1")

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.Key["SK"]; got != "TRANSACTION#transaction-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "TRANSACTION#transaction-1")
		}
	})
}

func TestDynamoDBTransactionRepository_Update(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	newClient := func(updateItem func(context.Context, *dynamodbclient.UpdateItemInput) error) *fakeDynamoDBClient {
		return &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findTransactionItem)
				*item = findTransactionItem{
					ID:         "transaction-1",
					Amount:     attributevalue.Number("10"),
					Type:       "EXPENSE",
					AccountID:  "account-1",
					CategoryID: "category-1",
					OccurredAt: timestamp,
					CreatedAt:  timestamp,
					UpdatedAt:  timestamp,
				}
				return nil
			},
			updateItem: updateItem,
		}
	}

	amount, err := valueobject.NewDecimal("20")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}

	transaction := entity.Transaction{
		ID:         "transaction-1",
		Amount:     amount,
		Type:       valueobject.TransactionTypeIncome,
		AccountID:  "account-2",
		CategoryID: "category-2",
		OccurredAt: updatedAt,
		UpdatedAt:  updatedAt,
	}

	t.Run("success", func(t *testing.T) {
		client := newClient(func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
			return nil
		})

		repository := NewDynamoDBTransactionRepository(client, "table")

		if got := repository.Update(context.Background(), "user-1", transaction); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("concurrent modification", func(t *testing.T) {
		client := newClient(func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
			return dynamodbclient.ErrConditionFailed
		})

		repository := NewDynamoDBTransactionRepository(client, "table")

		want := port.ErrConcurrentModification
		got := repository.Update(context.Background(), "user-1", transaction)

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("get current error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return boom
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		got := repository.Update(context.Background(), "user-1", entity.Transaction{})

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("unexpected update error", func(t *testing.T) {
		boom := errors.New("boom")

		client := newClient(func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
			return boom
		})

		repository := NewDynamoDBTransactionRepository(client, "table")

		got := repository.Update(context.Background(), "user-1", transaction)

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected update item", func(t *testing.T) {
		var captured *dynamodbclient.UpdateItemInput

		client := newClient(func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
			captured = input
			return nil
		})

		repository := NewDynamoDBTransactionRepository(client, "table")

		_ = repository.Update(context.Background(), "user-1", transaction)

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.Key["SK"]; got != "TRANSACTION#transaction-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "TRANSACTION#transaction-1")
		}

		if captured.ConditionExpression == nil || *captured.ConditionExpression != "updated_at = :old_updated_at" {
			t.Errorf("got %v, want %v", captured.ConditionExpression, "updated_at = :old_updated_at")
		}

		gsiSortKey := "TRANSACTION#" + updatedAt.Format(time.RFC3339Nano) + "#transaction-1"
		wantValues := map[string]any{
			":amount":         dynamodbclient.Decimal(amount),
			":type":           "INCOME",
			":account_id":     "account-2",
			":category_id":    "category-2",
			":occurred_at":    updatedAt.Format(time.RFC3339Nano),
			":updated_at":     updatedAt.Format(time.RFC3339Nano),
			":old_updated_at": createdAt.Format(time.RFC3339Nano),
			":gsi1sk":         gsiSortKey,
			":gsi2pk":         "USER#user-1#TYPE#INCOME",
			":gsi2sk":         gsiSortKey,
			":gsi3pk":         "USER#user-1#ACCOUNT#account-2",
			":gsi3sk":         gsiSortKey,
			":gsi4pk":         "USER#user-1#CATEGORY#category-2",
			":gsi4sk":         gsiSortKey,
		}

		for key, want := range wantValues {
			if got := captured.ExpressionAttributeValues[key]; got != want {
				t.Errorf("expression value %q: got %v, want %v", key, got, want)
			}
		}

		if captured.ExpressionAttributeNames["#type"] != "type" {
			t.Errorf("got %v, want %v", captured.ExpressionAttributeNames, map[string]string{"#type": "type"})
		}
	})
}

func TestDynamoDBTransactionRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		if got := repository.Delete(context.Background(), "user-1", "transaction-1"); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("transaction not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		want := port.ErrTransactionNotFound
		got := repository.Delete(context.Background(), "user-1", "transaction-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return boom
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		got := repository.Delete(context.Background(), "user-1", "transaction-1")

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected update item", func(t *testing.T) {
		var captured *dynamodbclient.UpdateItemInput

		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				captured = input
				return nil
			},
		}

		repository := NewDynamoDBTransactionRepository(client, "table")

		_ = repository.Delete(context.Background(), "user-1", "transaction-1")

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.Key["SK"]; got != "TRANSACTION#transaction-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "TRANSACTION#transaction-1")
		}

		if want := "SET deleted_at = :deleted_at, updated_at = :updated_at"; captured.UpdateExpression != want {
			t.Errorf("got %v, want %v", captured.UpdateExpression, want)
		}

		if captured.ConditionExpression == nil || *captured.ConditionExpression != "attribute_exists(PK)" {
			t.Errorf("got %v, want %v", captured.ConditionExpression, "attribute_exists(PK)")
		}

		deletedAt, ok := captured.ExpressionAttributeValues[":deleted_at"].(string)
		if !ok || deletedAt == "" {
			t.Errorf("got %v, want non-empty string", captured.ExpressionAttributeValues[":deleted_at"])
		}

		if captured.ExpressionAttributeValues[":updated_at"] != deletedAt {
			t.Errorf("got %v, want %v", captured.ExpressionAttributeValues[":updated_at"], deletedAt)
		}
	})
}

func TestFindTransactionItem_toDomain(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success", func(t *testing.T) {
		item := findTransactionItem{
			ID:          "transaction-1",
			Amount:      attributevalue.Number("10"),
			Type:        "EXPENSE",
			AccountID:   "account-1",
			CategoryID:  "category-1",
			Description: "Lunch",
			OccurredAt:  timestamp,
			CreatedAt:   timestamp,
			UpdatedAt:   timestamp,
		}

		got, err := item.toDomain()
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "transaction-1" {
			t.Errorf("got %v, want %v", got.ID, "transaction-1")
		}

		if got.Amount.String() != "10" {
			t.Errorf("got %v, want %v", got.Amount.String(), "10")
		}

		if got.Type != valueobject.TransactionTypeExpense {
			t.Errorf("got %v, want %v", got.Type, valueobject.TransactionTypeExpense)
		}

		if got.AccountID != "account-1" || got.CategoryID != "category-1" {
			t.Errorf("got %v/%v, want account-1/category-1", got.AccountID, got.CategoryID)
		}

		if got.Description != "Lunch" {
			t.Errorf("got %v, want %v", got.Description, "Lunch")
		}

		if !got.OccurredAt.Equal(createdAt) || !got.CreatedAt.Equal(createdAt) || !got.UpdatedAt.Equal(createdAt) {
			t.Errorf("got %v/%v/%v, want %v", got.OccurredAt, got.CreatedAt, got.UpdatedAt, createdAt)
		}

		if got.DeletedAt != nil {
			t.Errorf("got %v, want %v", got.DeletedAt, nil)
		}
	})

	t.Run("deleted at", func(t *testing.T) {
		item := findTransactionItem{
			Amount:     attributevalue.Number("10"),
			OccurredAt: timestamp,
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			DeletedAt:  &timestamp,
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

	t.Run("invalid amount", func(t *testing.T) {
		item := findTransactionItem{
			Amount:     attributevalue.Number("nope"),
			OccurredAt: timestamp,
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid occurred_at", func(t *testing.T) {
		item := findTransactionItem{
			Amount:     attributevalue.Number("10"),
			OccurredAt: "nope",
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid created_at", func(t *testing.T) {
		item := findTransactionItem{
			Amount:     attributevalue.Number("10"),
			OccurredAt: timestamp,
			CreatedAt:  "nope",
			UpdatedAt:  timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid updated_at", func(t *testing.T) {
		item := findTransactionItem{
			Amount:     attributevalue.Number("10"),
			OccurredAt: timestamp,
			CreatedAt:  timestamp,
			UpdatedAt:  "nope",
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid deleted_at", func(t *testing.T) {
		invalid := "nope"
		item := findTransactionItem{
			Amount:     attributevalue.Number("10"),
			OccurredAt: timestamp,
			CreatedAt:  timestamp,
			UpdatedAt:  timestamp,
			DeletedAt:  &invalid,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})
}
