package category

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBCategoryRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBCategoryRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBCategoryRepository {
	return &DynamoDBCategoryRepository{
		client:    client,
		tableName: tableName,
	}
}

func (r *DynamoDBCategoryRepository) Create(
	ctx context.Context,
	userID string,
	name string,
	transactionType valueobjects.TransactionType,
) error {
	id := uuid.New()
	now := time.Now().UTC()

	err := r.client.PutItem(ctx, r.tableName, map[string]any{
		"PK":         fmt.Sprintf("USER#%s", userID),
		"SK":         fmt.Sprintf("CATEGORY#%s", id),
		"name":       name,
		"type":       string(transactionType),
		"created_at": now.Format(time.RFC3339Nano),
		"updated_at": now.Format(time.RFC3339Nano),
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrItemExists) {
			return port.ErrCategoryExists
		}
	}

	return err
}
