package category

import (
	"context"
	"errors"
	"fmt"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBCategoryRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBCategoryRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBCategoryRepository {
	return &DynamoDBCategoryRepository{client: client, tableName: tableName}
}

func (r *DynamoDBCategoryRepository) Create(ctx context.Context, userID string, category entities.Category) error {
	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         fmt.Sprintf("USER#%s", userID),
					"SK":         fmt.Sprintf("CATEGORY#%s", category.ID),
					"name":       category.Name,
					"type":       string(category.Type),
					"created_at": category.CreatedAt.Format(time.RFC3339Nano),
					"updated_at": category.UpdatedAt.Format(time.RFC3339Nano),
				},
			},
		},
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":          fmt.Sprintf("USER#%s", userID),
					"SK":          fmt.Sprintf("CATEGORY_NAME#%s", category.Name),
					"category_id": category.ID,
				},
				Condition: "attribute_not_exists(PK)",
			},
		},
	}

	err := r.client.TransactWriteItems(ctx, transactItems)
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrItemExists) {
			return port.ErrCategoryExists
		}
	}

	return err
}
