package category

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/category/application/ports/categoryrepository"
	entities "github.com/hyoaru/itala-api/internal/features/category/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBCategoryRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBCategoryRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBCategoryRepository {
	return &DynamoDBCategoryRepository{client: client, tableName: tableName}
}

type findCategoryItem struct {
	ID              string `dynamodbav:"id"`
	Name            string `dynamodbav:"name"`
	TransactionType string `dynamodbav:"transaction_type"`
	CreatedAt       string `dynamodbav:"created_at"`
	UpdatedAt       string `dynamodbav:"updated_at"`
}

func (i findCategoryItem) toDomain() (entities.Category, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, i.CreatedAt)
	if err != nil {
		return entities.Category{}, fmt.Errorf("parse created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, i.UpdatedAt)
	if err != nil {
		return entities.Category{}, fmt.Errorf("parse updated_at: %w", err)
	}

	category := entities.Category{
		ID:        i.ID,
		Name:      i.Name,
		Type:      valueobjects.TransactionType(i.TransactionType),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return category, nil
}

func (r *DynamoDBCategoryRepository) Create(ctx context.Context, userID string, category entities.Category) error {
	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         fmt.Sprintf("USER#%s", userID),
					"SK":         fmt.Sprintf("CATEGORY#%s", category.ID),
					"id":         category.ID,
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

func (r *DynamoDBCategoryRepository) Find(ctx context.Context, userID string, query port.CategoryQuery) (port.CategoryPage, error) {
	conditionExpression := "PK = :pk AND begins_with(SK, :sk)"
	expressionValues := map[string]any{":pk": fmt.Sprintf("USER#%s", userID), ":sk": "ACCOUNT#"}

	var filters []string
	if query.Type != nil {
		filters = append(filters, "transaction_type = :type")
		expressionValues[":type"] = string(*query.Type)
	}
	if query.Name != nil {
		filters = append(filters, "name = :name")
		expressionValues[":name"] = *query.Name
	}
	filterExpression := strings.Join(filters, " AND ")

	var startKey map[string]any
	if query.Cursor != nil {
		decodedCursor, err := base64.RawURLEncoding.DecodeString(*query.Cursor)
		if err != nil {
			return port.CategoryPage{}, fmt.Errorf("decode cursor: %w", err)
		}
		if err := json.Unmarshal(decodedCursor, &startKey); err != nil {
			return port.CategoryPage{}, fmt.Errorf("unmarshal start key: %w", err)
		}
	}

	var queryItems []findCategoryItem
	metadata, err := r.client.Query(
		ctx,
		r.tableName,
		query.Limit,
		conditionExpression,
		filterExpression,
		expressionValues,
		startKey,
		&queryItems,
	)
	if err != nil {
		return port.CategoryPage{}, fmt.Errorf("find categories: %w", err)
	}

	categories := make([]entities.Category, 0, len(queryItems))
	for _, item := range queryItems {
		category, err := item.toDomain()
		if err != nil {
			return port.CategoryPage{}, err
		}
		categories = append(categories, category)
	}

	if metadata.LastEvaluatedKey == nil {
		return port.CategoryPage{Categories: categories, NextCursor: nil}, nil
	}

	encodedNextCursor, err := json.Marshal(metadata.LastEvaluatedKey)
	if err != nil {
		return port.CategoryPage{}, fmt.Errorf("marshal next cursor: %w", err)
	}

	nextCursor := base64.RawURLEncoding.EncodeToString(encodedNextCursor)
	return port.CategoryPage{Categories: categories, NextCursor: &nextCursor}, nil
}
