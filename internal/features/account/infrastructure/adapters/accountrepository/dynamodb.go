package account

import (
	"context"
	"errors"
	"fmt"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBAccountRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBAccountRepository {
	return &DynamoDBAccountRepository{client: client, tableName: tableName}
}

func (r *DynamoDBAccountRepository) Create(ctx context.Context, userID string, account entities.Account) error {
	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         fmt.Sprintf("USER#%s", userID),
					"SK":         fmt.Sprintf("ACCOUNT#%s", account.ID),
					"id":         account.ID,
					"name":       account.Name,
					"balance":    dynamodbclient.Decimal(account.Balance),
					"created_at": account.CreatedAt.Format(time.RFC3339Nano),
					"updated_at": account.UpdatedAt.Format(time.RFC3339Nano),
				},
			},
		},
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         fmt.Sprintf("USER#%s", userID),
					"SK":         fmt.Sprintf("ACCOUNT_NAME#%s", account.Name),
					"account_id": account.ID,
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
