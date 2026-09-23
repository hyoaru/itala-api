package idempotency

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"

	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type fakeDynamoDBClient struct {
	dynamodbclient.DynamoDBClient
	putItem    func(ctx context.Context, input *dynamodbclient.PutItemInput) error
	getItem    func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error
	updateItem func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error
	deleteItem func(ctx context.Context, input *dynamodbclient.DeleteItemInput) error
}

func (f *fakeDynamoDBClient) PutItem(ctx context.Context, input *dynamodbclient.PutItemInput) error {
	return f.putItem(ctx, input)
}

func (f *fakeDynamoDBClient) GetItem(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
	return f.getItem(ctx, input, output)
}

func (f *fakeDynamoDBClient) UpdateItem(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
	return f.updateItem(ctx, input)
}

func (f *fakeDynamoDBClient) DeleteItem(ctx context.Context, input *dynamodbclient.DeleteItemInput) error {
	return f.deleteItem(ctx, input)
}

func idempotencyPK(key string) string {
	return fmt.Sprintf("IDEMPOTENCY#%x", sha256.Sum256([]byte(key)))
}

func TestDynamoDBIdempotencyStore_Acquire(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		lock, status, result, err := store.Acquire(context.Background(), "key-1", 900)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if status != IdempotencyStatusAcquired {
			t.Errorf("got %v, want %v", status, IdempotencyStatusAcquired)
		}

		if result != "" {
			t.Errorf("got %v, want %v", result, "")
		}

		if lock.Key != "key-1" {
			t.Errorf("got %v, want %v", lock.Key, "key-1")
		}

		if lock.Token == "" {
			t.Error("got empty token, want non-empty")
		}
	})

	t.Run("condition failed completed", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*acquireItem)
				item.Status = string(IdempotencyStatusCompleted)
				item.Result = "result-1"
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		_, status, result, err := store.Acquire(context.Background(), "key-1", 900)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if status != IdempotencyStatusCompleted {
			t.Errorf("got %v, want %v", status, IdempotencyStatusCompleted)
		}

		if result != "result-1" {
			t.Errorf("got %v, want %v", result, "result-1")
		}
	})

	t.Run("condition failed locked", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*acquireItem)
				item.Status = string(IdempotencyStatusLocked)
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		_, status, result, err := store.Acquire(context.Background(), "key-1", 900)
		if err != nil {
			t.Fatalf("got %v, want %v", err, nil)
		}

		if status != IdempotencyStatusLocked {
			t.Errorf("got %v, want %v", status, IdempotencyStatusLocked)
		}

		if result != "" {
			t.Errorf("got %v, want %v", result, "")
		}
	})

	t.Run("condition failed invalid status", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				item := output.(*acquireItem)
				item.Status = "BOGUS"
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		if _, _, _, err := store.Acquire(context.Background(), "key-1", 900); err == nil {
			t.Errorf("got %v, want non-nil", err)
		}
	})

	t.Run("condition failed item not found", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return dynamodbclient.ErrItemNotFound
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		_, _, _, err := store.Acquire(context.Background(), "key-1", 900)

		if !errors.Is(err, ErrItemNotFound) {
			t.Errorf("got %v, want %v", err, ErrItemNotFound)
		}
	})

	t.Run("condition failed get error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
			getItem: func(ctx context.Context, input *dynamodbclient.GetItemInput, output any) error {
				return boom
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		_, _, _, err := store.Acquire(context.Background(), "key-1", 900)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("put error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			putItem: func(ctx context.Context, input *dynamodbclient.PutItemInput) error {
				return boom
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		_, _, _, err := store.Acquire(context.Background(), "key-1", 900)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
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

		store := NewDynamoDBIdempotencyStore(client, "table")

		lock, _, _, _ := store.Acquire(context.Background(), "key-1", 900)

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Item["PK"]; got != idempotencyPK("key-1") {
			t.Errorf("PK: got %v, want %v", got, idempotencyPK("key-1"))
		}

		if got := captured.Item["SK"]; got != "#LOCK" {
			t.Errorf("SK: got %v, want %v", got, "#LOCK")
		}

		if got := captured.Item["status"]; got != string(IdempotencyStatusLocked) {
			t.Errorf("status: got %v, want %v", got, string(IdempotencyStatusLocked))
		}

		if got := captured.Item["token"]; got != lock.Token || got == "" {
			t.Errorf("token: got %v, want %v", got, lock.Token)
		}

		expiresAt, ok := captured.Item["expires_at"].(int64)
		if !ok || expiresAt <= 0 {
			t.Errorf("expires_at: got %v, want positive int64", captured.Item["expires_at"])
		}

		if captured.ConditionExpression == nil {
			t.Fatal("got nil condition expression, want non-nil")
		}

		if captured.ExpressionAttributeNames["#expires_at"] != "expires_at" {
			t.Errorf("got %v, want #expires_at->expires_at", captured.ExpressionAttributeNames)
		}

		if _, ok := captured.ExpressionAttributeValues[":now"]; !ok {
			t.Errorf("got %v, want :now present", captured.ExpressionAttributeValues)
		}
	})
}

func TestDynamoDBIdempotencyStore_Commit(t *testing.T) {
	lock := IdempotencyLock{Key: "key-1", Token: "token-1"}

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		if err := store.Commit(context.Background(), lock, "result-1"); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
	})

	t.Run("condition failed", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		err := store.Commit(context.Background(), lock, "result-1")

		if !errors.Is(err, ErrInvalidLockToken) {
			t.Errorf("got %v, want %v", err, ErrInvalidLockToken)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			updateItem: func(ctx context.Context, input *dynamodbclient.UpdateItemInput) error {
				return boom
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		err := store.Commit(context.Background(), lock, "result-1")

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
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

		store := NewDynamoDBIdempotencyStore(client, "table")

		_ = store.Commit(context.Background(), lock, "result-1")

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != idempotencyPK("key-1") {
			t.Errorf("PK: got %v, want %v", got, idempotencyPK("key-1"))
		}

		if got := captured.Key["SK"]; got != "#LOCK" {
			t.Errorf("SK: got %v, want %v", got, "#LOCK")
		}

		if want := "SET #result = :result, #status = :completed"; captured.UpdateExpression != want {
			t.Errorf("got %v, want %v", captured.UpdateExpression, want)
		}

		if captured.ConditionExpression == nil || *captured.ConditionExpression != "#status = :locked AND #token = :token" {
			t.Errorf("got %v, want %v", captured.ConditionExpression, "#status = :locked AND #token = :token")
		}

		wantValues := map[string]any{
			":result":    "result-1",
			":locked":    IdempotencyStatusLocked,
			":token":     "token-1",
			":completed": IdempotencyStatusCompleted,
		}

		for key, want := range wantValues {
			if got := captured.ExpressionAttributeValues[key]; got != want {
				t.Errorf("expression value %q: got %v, want %v", key, got, want)
			}
		}

		wantNames := map[string]string{"#status": "status", "#token": "token", "#result": "result"}
		for key, want := range wantNames {
			if got := captured.ExpressionAttributeNames[key]; got != want {
				t.Errorf("expression name %q: got %v, want %v", key, got, want)
			}
		}
	})
}

func TestDynamoDBIdempotencyStore_Release(t *testing.T) {
	lock := IdempotencyLock{Key: "key-1", Token: "token-1"}

	t.Run("success", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			deleteItem: func(ctx context.Context, input *dynamodbclient.DeleteItemInput) error {
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		if err := store.Release(context.Background(), lock); err != nil {
			t.Errorf("got %v, want %v", err, nil)
		}
	})

	t.Run("condition failed", func(t *testing.T) {
		client := &fakeDynamoDBClient{
			deleteItem: func(ctx context.Context, input *dynamodbclient.DeleteItemInput) error {
				return dynamodbclient.ErrConditionFailed
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		err := store.Release(context.Background(), lock)

		if !errors.Is(err, ErrInvalidLockToken) {
			t.Errorf("got %v, want %v", err, ErrInvalidLockToken)
		}
	})

	t.Run("unexpected error", func(t *testing.T) {
		boom := errors.New("boom")

		client := &fakeDynamoDBClient{
			deleteItem: func(ctx context.Context, input *dynamodbclient.DeleteItemInput) error {
				return boom
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		err := store.Release(context.Background(), lock)

		if !errors.Is(err, boom) {
			t.Errorf("got %v, want %v", err, boom)
		}
	})

	t.Run("writes expected delete item", func(t *testing.T) {
		var captured *dynamodbclient.DeleteItemInput

		client := &fakeDynamoDBClient{
			deleteItem: func(ctx context.Context, input *dynamodbclient.DeleteItemInput) error {
				captured = input
				return nil
			},
		}

		store := NewDynamoDBIdempotencyStore(client, "table")

		_ = store.Release(context.Background(), lock)

		if captured.TableName != "table" {
			t.Errorf("got %v, want %v", captured.TableName, "table")
		}

		if got := captured.Key["PK"]; got != idempotencyPK("key-1") {
			t.Errorf("PK: got %v, want %v", got, idempotencyPK("key-1"))
		}

		if got := captured.Key["SK"]; got != "#LOCK" {
			t.Errorf("SK: got %v, want %v", got, "#LOCK")
		}

		if captured.ConditionExpression == nil || *captured.ConditionExpression != "#status = :locked AND #token = :token" {
			t.Errorf("got %v, want %v", captured.ConditionExpression, "#status = :locked AND #token = :token")
		}

		if got := captured.ExpressionAttributeValues[":token"]; got != "token-1" {
			t.Errorf(":token got %v, want %v", got, "token-1")
		}

		if got := captured.ExpressionAttributeValues[":locked"]; got != IdempotencyStatusLocked {
			t.Errorf(":locked got %v, want %v", got, IdempotencyStatusLocked)
		}

		if captured.ExpressionAttributeNames["#status"] != "status" || captured.ExpressionAttributeNames["#token"] != "token" {
			t.Errorf("got %v, want #status->status,#token->token", captured.ExpressionAttributeNames)
		}
	})
}
