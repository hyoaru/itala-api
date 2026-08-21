package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/shopspring/decimal"

	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type dynamodbDecimal decimal.Decimal

func (d dynamodbDecimal) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return &types.AttributeValueMemberN{
		Value: decimal.Decimal(d).String(),
	}, nil
}

type DynamoDBTransactionRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBTransactionRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBTransactionRepository {
	return &DynamoDBTransactionRepository{client: client, tableName: tableName}
}

func (r *DynamoDBTransactionRepository) Create(ctx context.Context, userID string, transaction entities.Transaction) error {
	return r.client.PutItem(ctx, r.tableName, map[string]any{
		"PK":               fmt.Sprintf("USER#%s", userID),
		"SK":               fmt.Sprintf("TRANSACTION#%s#%s", transaction.OccurredAt.Format(time.RFC3339Nano), transaction.ID),
		"amount":           dynamodbDecimal(transaction.Amount),
		"transaction_type": string(transaction.Type),
		"category_id":      transaction.CategoryID,
		"description":      transaction.Description,
		"occurred_at":      transaction.OccurredAt.Format(time.RFC3339Nano),
		"created_at":       transaction.CreatedAt.Format(time.RFC3339Nano),
		"updated_at":       transaction.UpdatedAt.Format(time.RFC3339Nano),
	})
}
