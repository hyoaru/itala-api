package idempotency

import (
	"context"
	"encoding/hex"
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

func (i *DynamoDBIdempotencyStore) Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, *IdempotencyStatus, *ResultJSON, error) {
	pk := fmt.Sprintf("IDEMPOTENCY#%s", hex.EncodeToString([]byte(key)))
	sk := "#LOCK"
	now := time.Now().UTC()
	token := uuid.New().String()

	ttlTimestamp := now.Add(time.Duration(ttl) * time.Second)

	err := i.client.PutItem(ctx, &dynamodbclient.PutItemInput{
		TableName: i.tableName,
		Item: map[string]any{
			"PK":     pk,
			"SK":     sk,
			"ttl":    ttlTimestamp.Unix(),
			"status": string(IdempotencyStatusLocked),
			"token":  token,
		},
		ConditionExpression:       aws.String("attribute_not_exists(PK) or ttl <= :now"),
		ExpressionAttributeValues: map[string]any{":now": now.Unix()},
	})

	if err == nil {
		s := IdempotencyStatusAcquired
		return IdempotencyLock{Key: key, Token: token}, &s, nil, nil
	}

	if !errors.Is(err, dynamodbclient.ErrConditionFailed) {
		return IdempotencyLock{}, nil, nil, fmt.Errorf("acquire idempotency: %w", err)
	}

	var item acquireItem
	err = i.client.GetItem(ctx, &dynamodbclient.GetItemInput{TableName: i.tableName, Key: map[string]any{"PK": pk, "SK": sk}}, &item)
	if err != nil {
		if !errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return IdempotencyLock{}, nil, nil, fmt.Errorf("get item: %w", err)
		}

		return IdempotencyLock{}, nil, nil, ErrItemNotFound
	}

	switch IdempotencyStatus(item.Status) {
	case IdempotencyStatusCompleted:
		s := IdempotencyStatusCompleted
		r := ResultJSON(item.Result)
		return IdempotencyLock{}, &s, &r, nil
	case IdempotencyStatusLocked:
		s := IdempotencyStatusLocked
		return IdempotencyLock{}, &s, nil, nil
	default:
		return IdempotencyLock{}, nil, nil, fmt.Errorf("invalid idempotency status: %q", item.Status)
	}
}

func (i *DynamoDBIdempotencyStore) Commit(ctx context.Context, lock IdempotencyLock, result string) error {
	pk := fmt.Sprintf("IDEMPOTENCY#%s", hex.EncodeToString([]byte(lock.Key)))
	sk := "#LOCK"

	conditionExpression := "status = :locked AND token = :token"
	err := i.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:           i.tableName,
		Key:                 map[string]any{"PK": pk, "SK": sk},
		UpdateExpression:    "SET result = :result, status = :completed",
		ConditionExpression: &conditionExpression,
		ExpressionAttributeValues: map[string]any{
			":result":    result,
			":locked":    IdempotencyStatusLocked,
			":token":     lock.Token,
			":completed": IdempotencyStatusCompleted,
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
	pk := fmt.Sprintf("IDEMPOTENCY#%s", hex.EncodeToString([]byte(lock.Key)))
	sk := "#LOCK"

	conditionExpression := "status = :locked AND token = :token"
	err := i.client.DeleteItem(ctx, &dynamodbclient.DeleteItemInput{
		TableName:                 i.tableName,
		Key:                       map[string]any{"PK": pk, "SK": sk},
		ConditionExpression:       &conditionExpression,
		ExpressionAttributeValues: map[string]any{":token": lock.Token, ":locked": IdempotencyStatusLocked},
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return ErrInvalidLockToken
		}

		return fmt.Errorf("release idempotency: %w", err)
	}

	return nil
}
