package idempotency

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
}

func (i *DynamoDBIdempotencyStore) Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, error) {
	pk := fmt.Sprintf("IDEMPOTENCY#%s", hex.EncodeToString([]byte(key)))
	sk := "#LOCK"
	now := time.Now().UTC()
	ttlTimestamp := now.Add(time.Duration(ttl) * time.Second)

	err := i.client.PutItem(ctx, &dynamodbclient.PutItemInput{
		TableName: i.tableName,
		Item: map[string]any{
			"PK":     pk,
			"SK":     sk,
			"status": string(IdempotencyStatusLocked),
			"ttl":    ttlTimestamp.Unix(),
		},
		ConditionExpression:       aws.String("attribute_not_exists(PK) or ttl <= :now"),
		ExpressionAttributeValues: map[string]any{":now": now.Unix()},
	})

	if err == nil {
		return IdempotencyLock{Status: IdempotencyStatusAcquired, Result: nil}, nil
	}

	if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); !ok {
		return IdempotencyLock{}, fmt.Errorf("acquire: %w", err)
	}

	var item acquireItem
	if err := i.client.GetItem(ctx, &dynamodbclient.GetItemInput{
		TableName: i.tableName,
		Key:       map[string]any{"PK": pk, "SK": sk},
	}, &item); err != nil {
		return IdempotencyLock{}, fmt.Errorf("get item: %w", err)
	}

	switch IdempotencyStatus(item.Status) {
	case IdempotencyStatusCompleted:
		return IdempotencyLock{Status: IdempotencyStatusCompleted, Result: item.Result}, nil
	case IdempotencyStatusLocked:
		return IdempotencyLock{Status: IdempotencyStatusLocked, Result: nil}, nil
	default:
		return IdempotencyLock{}, fmt.Errorf("invalid idempotency status: %q", item.Status)
	}
}

func (i *DynamoDBIdempotencyStore) Commit(ctx context.Context, key string, result any) error {
	return nil
}

func (i *DynamoDBIdempotencyStore) Release(ctx context.Context, key string) error {
	return nil
}
