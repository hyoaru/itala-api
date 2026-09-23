package account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type fakeDynamoDBClient struct {
	dynamodbclient.DynamoDBClient
	transactWriteItems func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error
	query              func(ctx context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error)
	getItem            func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error
	updateItem         func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error
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

func (f *fakeDynamoDBClient) UpdateItem(
	ctx context.Context,
	input *dynamodbclient.UpdateItemInput,
) error {
	return f.updateItem(ctx, input)
}

func TestDynamoDBAccountRepository_Create(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	balance, err := valueobject.NewDecimal("100")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		if got := repository.Create(context.Background(), "user-1", entity.Account{}); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("account already exists", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		want := port.ErrAccountExists
		got := repository.Create(context.Background(), "user-1", entity.Account{})

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				return boom
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		got := repository.Create(context.Background(), "user-1", entity.Account{})

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected transaction items", func(t *testing.T) {
		var captured *dynamodbclient.TransactWriteItemsInput

		client := &fakeDynamoDBClient{
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				captured = input
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		account := entity.Account{
			ID:        "account-1",
			Name:      "Savings",
			Balance:   balance,
			CreatedAt: createdAt,
			UpdatedAt: createdAt,
		}

		_ = repository.Create(context.Background(), "user-1", account)

		if len(captured.TransactItems) != 2 {
			t.Fatalf("got %d transact items, want %d", len(captured.TransactItems), 2)
		}

		accountPut := captured.TransactItems[0].Put
		if accountPut == nil {
			t.Fatal("got nil account put, want non-nil")
		}

		if accountPut.TableName != "table" {
			t.Errorf("got %v, want %v", accountPut.TableName, "table")
		}

		if accountPut.ConditionExpression != nil {
			t.Errorf("got %v, want %v", accountPut.ConditionExpression, nil)
		}

		wantAccountItem := map[string]any{
			"PK":         "USER#user-1",
			"SK":         "ACCOUNT#account-1",
			"id":         "account-1",
			"name":       "Savings",
			"balance":    dynamodbclient.Decimal(balance),
			"created_at": createdAt.Format(time.RFC3339Nano),
			"updated_at": createdAt.Format(time.RFC3339Nano),
		}

		for key, want := range wantAccountItem {
			if got := accountPut.Item[key]; got != want {
				t.Errorf("account item %q: got %v, want %v", key, got, want)
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
			"PK":         "USER#user-1",
			"SK":         "ACCOUNT_NAME#Savings",
			"account_id": "account-1",
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

func TestDynamoDBAccountRepository_Find(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			query: func(
				ctx context.Context,
				input *dynamodbclient.QueryInput,
				output any,
			) (dynamodbclient.QueryOutput, error) {
				items := output.(*[]findAccountItem)
				*items = append(*items, findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					CreatedAt: createdAt.Format(time.RFC3339Nano),
					UpdatedAt: createdAt.Format(time.RFC3339Nano),
				})
				return dynamodbclient.QueryOutput{}, nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		page, err := repository.Find(context.Background(), "user-1", port.AccountQuery{})
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if len(page.Accounts) != 1 {
			t.Fatalf("got %d accounts, want %d", len(page.Accounts), 1)
		}

		got := page.Accounts[0]
		if got.ID != "account-1" {
			t.Errorf("got %v, want %v", got.ID, "account-1")
		}

		if got.Name != "Savings" {
			t.Errorf("got %v, want %v", got.Name, "Savings")
		}

		if got.Balance.String() != "100" {
			t.Errorf("got %v, want %v", got.Balance.String(), "100")
		}

		if !got.CreatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.CreatedAt, createdAt)
		}

		if !got.UpdatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, createdAt)
		}
	})

	t.Run("success with next cursor", func(t *testing.T) {
		lastEvaluatedKey := map[string]any{"PK": "USER#user-1", "SK": "ACCOUNT#account-1"}

		client := &fakeDynamoDBClient{
			query: func(
				ctx context.Context,
				input *dynamodbclient.QueryInput,
				output any,
			) (dynamodbclient.QueryOutput, error) {
				return dynamodbclient.QueryOutput{LastEvaluatedKey: lastEvaluatedKey}, nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		page, err := repository.Find(context.Background(), "user-1", port.AccountQuery{})
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

		repository := NewDynamoDBAccountRepository(client, "table")

		_, err := repository.Find(context.Background(), "user-1", port.AccountQuery{})
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

		repository := NewDynamoDBAccountRepository(client, "table")

		name := "Savings"
		cursor := base64.RawURLEncoding.EncodeToString([]byte(`{"PK":"USER#user-1","SK":"ACCOUNT#account-1"}`))

		query := port.AccountQuery{Limit: 10, Name: &name, Cursor: &cursor}

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

		wantFilter := "(attribute_not_exists(deleted_at) OR attribute_type(deleted_at, :nullType)) AND #name = :name"
		if captured.FilterExpression == nil || *captured.FilterExpression != wantFilter {
			t.Errorf("got %v, want %v", captured.FilterExpression, wantFilter)
		}

		wantValues := map[string]any{
			":pk":       "USER#user-1",
			":sk":       "ACCOUNT#",
			":nullType": "NULL",
			":name":     "Savings",
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

		if got := captured.ExclusiveStartKey["SK"]; got != "ACCOUNT#account-1" {
			t.Errorf("start key %q: got %v, want %v", "SK", got, "ACCOUNT#account-1")
		}
	})
}

func TestDynamoDBAccountRepository_FindOne(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		got, err := repository.FindOne(context.Background(), "user-1", "account-1")
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "account-1" {
			t.Errorf("got %v, want %v", got.ID, "account-1")
		}

		if got.Name != "Savings" {
			t.Errorf("got %v, want %v", got.Name, "Savings")
		}

		if got.Balance.String() != "100" {
			t.Errorf("got %v, want %v", got.Balance.String(), "100")
		}

		if !got.CreatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.CreatedAt, createdAt)
		}

		if !got.UpdatedAt.Equal(createdAt) {
			t.Errorf("got %v, want %v", got.UpdatedAt, createdAt)
		}
	})

	t.Run("account not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return dynamodbclient.ErrItemNotFound
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		want := port.ErrAccountNotFound
		_, got := repository.FindOne(context.Background(), "user-1", "account-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("account deleted", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					DeletedAt: &timestamp,
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		want := port.ErrAccountNotFound
		_, got := repository.FindOne(context.Background(), "user-1", "account-1")

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

		repository := NewDynamoDBAccountRepository(client, "table")

		_, got := repository.FindOne(context.Background(), "user-1", "account-1")

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

		repository := NewDynamoDBAccountRepository(client, "table")

		_, _ = repository.FindOne(context.Background(), "user-1", "account-1")

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.Key["SK"]; got != "ACCOUNT#account-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "ACCOUNT#account-1")
		}
	})
}

func TestDynamoDBAccountRepository_Update(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 3, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success same name", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		account := entity.Account{ID: "account-1", Name: "Savings", UpdatedAt: updatedAt}

		if got := repository.Update(context.Background(), "user-1", account); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("success name changed", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Old",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		account := entity.Account{ID: "account-1", Name: "Savings", UpdatedAt: updatedAt}

		if got := repository.Update(context.Background(), "user-1", account); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("concurrent modification", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		account := entity.Account{ID: "account-1", Name: "Savings", UpdatedAt: updatedAt}

		want := port.ErrConcurrentModification
		got := repository.Update(context.Background(), "user-1", account)

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("account already exists", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Old",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		account := entity.Account{ID: "account-1", Name: "Savings", UpdatedAt: updatedAt}

		want := port.ErrAccountExists
		got := repository.Update(context.Background(), "user-1", account)

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

		repository := NewDynamoDBAccountRepository(client, "table")

		got := repository.Update(context.Background(), "user-1", entity.Account{})

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("unexpected update error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return boom
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		account := entity.Account{ID: "account-1", Name: "Savings", UpdatedAt: updatedAt}

		got := repository.Update(context.Background(), "user-1", account)

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})
}

func TestDynamoDBAccountRepository_Delete(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	newClient := func(transactErr error) *fakeDynamoDBClient {
		return &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*findAccountItem)
				*item = findAccountItem{
					ID:        "account-1",
					Name:      "Savings",
					Balance:   attributevalue.Number("100"),
					CreatedAt: timestamp,
					UpdatedAt: timestamp,
				}
				return nil
			},
			transactWriteItems: func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
				return transactErr
			},
		}
	}

	t.Run("success", func(t *testing.T) {
		repository := NewDynamoDBAccountRepository(newClient(nil), "table")

		if got := repository.Delete(context.Background(), "user-1", "account-1"); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("account not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return dynamodbclient.ErrItemNotFound
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		want := port.ErrAccountNotFound
		got := repository.Delete(context.Background(), "user-1", "account-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("condition failed", func(t *testing.T) {
		repository := NewDynamoDBAccountRepository(newClient(dynamodbclient.ErrConditionFailed), "table")

		want := port.ErrAccountNotFound
		got := repository.Delete(context.Background(), "user-1", "account-1")

		if !errors.Is(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		repository := NewDynamoDBAccountRepository(newClient(boom), "table")

		got := repository.Delete(context.Background(), "user-1", "account-1")

		if !errors.Is(got, boom) {
			t.Errorf("got %v, want %v", got, boom)
		}
	})

	t.Run("writes expected transaction items", func(t *testing.T) {
		var captured *dynamodbclient.TransactWriteItemsInput

		client := newClient(nil)
		client.transactWriteItems = func(ctx context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
			captured = input
			return nil
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		_ = repository.Delete(context.Background(), "user-1", "account-1")

		if len(captured.TransactItems) != 2 {
			t.Fatalf("got %d transact items, want %d", len(captured.TransactItems), 2)
		}

		accountUpdate := captured.TransactItems[0].Update
		if accountUpdate == nil {
			t.Fatal("got nil account update, want non-nil")
		}

		if accountUpdate.TableName != "table" {
			t.Errorf("got %v, want %v", accountUpdate.TableName, "table")
		}

		if got := accountUpdate.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := accountUpdate.Key["SK"]; got != "ACCOUNT#account-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "ACCOUNT#account-1")
		}

		if want := "SET deleted_at = :deleted_at, updated_at = :updated_at"; accountUpdate.UpdateExpression != want {
			t.Errorf("got %v, want %v", accountUpdate.UpdateExpression, want)
		}

		if accountUpdate.ConditionExpression == nil {
			t.Fatal("got nil condition expression, want non-nil")
		}

		if want := "attribute_exists(PK)"; *accountUpdate.ConditionExpression != want {
			t.Errorf("got %v, want %v", *accountUpdate.ConditionExpression, want)
		}

		deletedAt, ok := accountUpdate.ExpressionAttributeValues[":deleted_at"].(string)
		if !ok || deletedAt == "" {
			t.Errorf("got %v, want non-empty string", accountUpdate.ExpressionAttributeValues[":deleted_at"])
		}

		if accountUpdate.ExpressionAttributeValues[":updated_at"] != deletedAt {
			t.Errorf("got %v, want %v", accountUpdate.ExpressionAttributeValues[":updated_at"], deletedAt)
		}

		nameDelete := captured.TransactItems[1].Delete
		if nameDelete == nil {
			t.Fatal("got nil name delete, want non-nil")
		}

		if nameDelete.TableName != "table" {
			t.Errorf("got %v, want %v", nameDelete.TableName, "table")
		}

		if got := nameDelete.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := nameDelete.Key["SK"]; got != "ACCOUNT_NAME#Savings" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "ACCOUNT_NAME#Savings")
		}
	})
}

func TestDynamoDBAccountRepository_AdjustBalance(t *testing.T) {
	delta, err := valueobject.NewDecimal("50")
	if err != nil {
		t.Fatalf("new decimal: %v", err)
	}

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return nil
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		if got := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta); got != nil {
			t.Errorf("got %v, want %v", got, nil)
		}
	})

	t.Run("account not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		repository := NewDynamoDBAccountRepository(client, "table")

		want := port.ErrAccountNotFound
		got := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

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

		repository := NewDynamoDBAccountRepository(client, "table")

		got := repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

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

		repository := NewDynamoDBAccountRepository(client, "table")

		_ = repository.AdjustBalance(context.Background(), "user-1", "account-1", "key-1", delta)

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != "USER#user-1" {
			t.Errorf("key %q: got %v, want %v", "PK", got, "USER#user-1")
		}

		if got := captured.Key["SK"]; got != "ACCOUNT#account-1" {
			t.Errorf("key %q: got %v, want %v", "SK", got, "ACCOUNT#account-1")
		}

		if want := "SET balance = balance + :delta, updated_at = :updated_at"; captured.UpdateExpression != want {
			t.Errorf("got %v, want %v", captured.UpdateExpression, want)
		}

		if captured.ConditionExpression == nil {
			t.Fatal("got nil condition expression, want non-nil")
		}

		if want := "attribute_exists(PK)"; *captured.ConditionExpression != want {
			t.Errorf("got %v, want %v", *captured.ConditionExpression, want)
		}

		if got := captured.ExpressionAttributeValues[":delta"]; got != dynamodbclient.Decimal(delta) {
			t.Errorf("delta: got %v, want %v", got, dynamodbclient.Decimal(delta))
		}

		updatedAt, ok := captured.ExpressionAttributeValues[":updated_at"].(string)
		if !ok || updatedAt == "" {
			t.Errorf("got %v, want non-empty string", captured.ExpressionAttributeValues[":updated_at"])
		}
	})
}

func TestFindAccountItem_toDomain(t *testing.T) {
	createdAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	timestamp := createdAt.Format(time.RFC3339Nano)

	t.Run("success", func(t *testing.T) {
		item := findAccountItem{
			ID:        "account-1",
			Name:      "Savings",
			Balance:   attributevalue.Number("100"),
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		got, err := item.toDomain()
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if got.ID != "account-1" {
			t.Errorf("got %v, want %v", got.ID, "account-1")
		}

		if got.Name != "Savings" {
			t.Errorf("got %v, want %v", got.Name, "Savings")
		}

		if got.Balance.String() != "100" {
			t.Errorf("got %v, want %v", got.Balance.String(), "100")
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
		item := findAccountItem{
			ID:        "account-1",
			Name:      "Savings",
			Balance:   attributevalue.Number("100"),
			DeletedAt: &timestamp,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
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

	t.Run("invalid balance", func(t *testing.T) {
		item := findAccountItem{
			Balance:   attributevalue.Number("nope"),
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid created_at", func(t *testing.T) {
		item := findAccountItem{
			Balance:   attributevalue.Number("100"),
			CreatedAt: "nope",
			UpdatedAt: timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid updated_at", func(t *testing.T) {
		item := findAccountItem{
			Balance:   attributevalue.Number("100"),
			CreatedAt: timestamp,
			UpdatedAt: "nope",
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("invalid deleted_at", func(t *testing.T) {
		invalid := "nope"
		item := findAccountItem{
			Balance:   attributevalue.Number("100"),
			DeletedAt: &invalid,
			CreatedAt: timestamp,
			UpdatedAt: timestamp,
		}

		if _, err := item.toDomain(); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})
}
