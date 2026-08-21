package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
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

func (r *DynamoDBTransactionRepository) Create(
	ctx context.Context,
	userID string,
	amount decimal.Decimal,
	transactionType valueobjects.TransactionType,
	categoryID string,
	description string,
	occurredAt time.Time,
) error {
	id := uuid.New()
	now := time.Now().UTC()
	occurredAt = occurredAt.UTC()

	return r.client.PutItem(ctx, r.tableName, map[string]any{
		"PK":               fmt.Sprintf("USER#%s", userID),
		"SK":               fmt.Sprintf("TRANSACTION#%s#%s", occurredAt.Format(time.RFC3339Nano), id),
		"amount":           dynamodbDecimal(amount),
		"transaction_type": string(transactionType),
		"category_id":      categoryID,
		"description":      description,
		"occurred_at":      occurredAt.UTC().Format(time.RFC3339Nano),
		"created_at":       now.Format(time.RFC3339Nano),
		"updated_at":       now.Format(time.RFC3339Nano),
	})
}
