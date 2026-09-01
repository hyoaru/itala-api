package idempotency

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBIdempotencyStore struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBIdempotencyStore(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBIdempotencyStore {
	return &DynamoDBIdempotencyStore{client: client, tableName: tableName}
}

type acquireItem struct {
	Status string `dynamodbav:"status"`
	Result string `dynamodbav:"result"`
	Token  string `dynamodbav:"token"`
}

func (i *DynamoDBIdempotencyStore) Acquire(ctx context.Context, key string, expiresAt uint16) (IdempotencyLock, IdempotencyStatus, ResultJSON, error) {
	pk := fmt.Sprintf("IDEMPOTENCY#%x", sha256.Sum256([]byte(key)))
	sk := "#LOCK"
	now := time.Now().UTC()
	token := uuid.New().String()

	expiresTimestamp := now.Add(time.Duration(expiresAt) * time.Second)

	err := i.client.PutItem(ctx, &dynamodbclient.PutItemInput{
		TableName: i.tableName,
		Item: map[string]any{
			"PK":         pk,
			"SK":         sk,
			"expires_at": expiresTimestamp.Unix(),
			"status":     string(IdempotencyStatusLocked),
			"token":      token,
		},
		ConditionExpression:       aws.String("attribute_not_exists(PK) or #expires_at <= :now"),
		ExpressionAttributeNames:  map[string]string{"#expires_at": "expires_at"},
		ExpressionAttributeValues: map[string]any{":now": now.Unix()},
	})

	if err == nil {
		return IdempotencyLock{Key: key, Token: token}, IdempotencyStatusAcquired, "", nil
	}

	if !errors.Is(err, dynamodbclient.ErrConditionFailed) {
		return IdempotencyLock{}, "", "", fmt.Errorf("acquire idempotency: %w", err)
	}

	var item acquireItem
	err = i.client.GetItem(ctx, &dynamodbclient.GetItemInput{TableName: i.tableName, Key: map[string]any{"PK": pk, "SK": sk}}, &item)
	if err != nil {
		if !errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return IdempotencyLock{}, "", "", fmt.Errorf("get existing idempotency item: %w", err)
		}

		return IdempotencyLock{}, "", "", ErrItemNotFound
	}

	switch IdempotencyStatus(item.Status) {
	case IdempotencyStatusCompleted:
		return IdempotencyLock{}, IdempotencyStatusCompleted, ResultJSON(item.Result), nil
	case IdempotencyStatusLocked:
		return IdempotencyLock{}, IdempotencyStatusLocked, "", nil
	default:
		return IdempotencyLock{}, "", "", fmt.Errorf("invalid idempotency status: %q", item.Status)
	}
}

func (i *DynamoDBIdempotencyStore) Commit(ctx context.Context, lock IdempotencyLock, result string) error {
	pk := fmt.Sprintf("IDEMPOTENCY#%x", sha256.Sum256([]byte(lock.Key)))
	sk := "#LOCK"

	conditionExpression := "#status = :locked AND #token = :token"
	err := i.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:           i.tableName,
		Key:                 map[string]any{"PK": pk, "SK": sk},
		UpdateExpression:    "SET #result = :result, #status = :completed",
		ConditionExpression: &conditionExpression,
		ExpressionAttributeValues: map[string]any{
			":result":    result,
			":locked":    IdempotencyStatusLocked,
			":token":     lock.Token,
			":completed": IdempotencyStatusCompleted,
		},
		ExpressionAttributeNames: map[string]string{
			"#status": "status",
			"#token":  "token",
			"#result": "result",
		},
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return ErrInvalidLockToken
		}

		return fmt.Errorf("commit idempotency: %w", err)
	}

	return nil
}

func (i *DynamoDBIdempotencyStore) Release(ctx context.Context, lock IdempotencyLock) error {
	pk := fmt.Sprintf("IDEMPOTENCY#%x", sha256.Sum256([]byte(lock.Key)))
	sk := "#LOCK"

	conditionExpression := "#status = :locked AND #token = :token"
	err := i.client.DeleteItem(ctx, &dynamodbclient.DeleteItemInput{
		TableName:                 i.tableName,
		Key:                       map[string]any{"PK": pk, "SK": sk},
		ConditionExpression:       &conditionExpression,
		ExpressionAttributeValues: map[string]any{":token": lock.Token, ":locked": IdempotencyStatusLocked},
		ExpressionAttributeNames:  map[string]string{"#status": "status", "#token": "token"},
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return ErrInvalidLockToken
		}

		return fmt.Errorf("release idempotency: %w", err)
	}

	return nil
}
