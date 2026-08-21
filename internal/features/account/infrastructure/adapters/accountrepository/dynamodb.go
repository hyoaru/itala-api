package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	port "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBAccountRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBAccountRepository {
	return &DynamoDBAccountRepository{client: client, tableName: tableName}
}

func (r *DynamoDBAccountRepository) Create(ctx context.Context, userID string, name string) error {
	id := uuid.New()
	now := time.Now().UTC()

	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         fmt.Sprintf("USER#%s", userID),
					"SK":         fmt.Sprintf("ACCOUNT#%s", id),
					"name":       name,
					"balance":    0,
					"created_at": now.Format(time.RFC3339Nano),
					"updated_at": now.Format(time.RFC3339Nano),
				},
			},
		},
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         fmt.Sprintf("USER#%s", userID),
					"SK":         fmt.Sprintf("ACCOUNT_NAME#%s", name),
					"account_id": id,
				},
				Condition: "attribute_not_exists(PK)",
			},
		},
	}

	err := r.client.TransactWriteItems(ctx, transactItems)
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrItemExists) {
			return port.ErrAccountExists
		}
	}

	return err
}
