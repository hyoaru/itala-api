package idempotency

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

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
	Status   string    `dynamodbav:"status"`
	Result   string    `dynamodbav:"result"`
	LockedAt time.Time `dynamodbav:"locked_at"`
}

func (i *DynamoDBIdempotencyStore) Acquire(ctx context.Context, key string, ttl uint16) (IdempotencyLock, error) {
	parts := strings.SplitN(key, ":", 2)
	pk := fmt.Sprintf("IDEMPOTENCY#%s", hex.EncodeToString([]byte(parts[0])))
	sk := fmt.Sprintf("#%s", parts[1])

	err := i.client.PutItem(ctx, &dynamodbclient.PutItemInput{
		TableName: i.tableName,
		Item:      map[string]any{"PK": pk, "SK": sk},
	})

	if err == nil {
		return IdempotencyLock{Status: IdempotencyStatusAcquired, Result: nil}, nil
	}

	if _, ok := errors.AsType[*types.ConditionalCheckFailedException](err); !ok {
		return IdempotencyLock{Status: IdempotencyStatusLocked, Result: nil}, nil
	}

	var item acquireItem
	if err = i.client.GetItem(ctx, &dynamodbclient.GetItemInput{
		TableName: i.tableName,
		Key:       map[string]any{"PK": pk, "SK": sk},
		Output:    &item,
	}); err != nil {
		return IdempotencyLock{}, fmt.Errorf("get item: %w", err)
	}

	return IdempotencyLock{Status: IdempotencyStatusCompleted, Result: item.Result}, nil
}

func (i *DynamoDBIdempotencyStore) Commit(ctx context.Context, key string, ttl uint16) error {
	return nil
}

func (i *DynamoDBIdempotencyStore) Release(ctx context.Context, key string, ttl uint16) error {
	return nil
}
