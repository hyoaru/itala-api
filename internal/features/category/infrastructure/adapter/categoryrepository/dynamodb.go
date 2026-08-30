package category

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	port "github.com/hyoaru/itala-api/internal/features/category/application/port/categoryrepository"
	entity "github.com/hyoaru/itala-api/internal/features/category/domain/entity"
	categoryvalueobject "github.com/hyoaru/itala-api/internal/features/category/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
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
	Status          string `dynamodbav:"status"`
	CreatedAt       string `dynamodbav:"created_at"`
	UpdatedAt       string `dynamodbav:"updated_at"`
}

func (i findCategoryItem) toDomain() (entity.Category, error) {
	createdAt, err := time.Parse(time.RFC3339Nano, i.CreatedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("parse created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, i.UpdatedAt)
	if err != nil {
		return entity.Category{}, fmt.Errorf("parse updated_at: %w", err)
	}

	category := entity.Category{
		ID:              i.ID,
		Name:            i.Name,
		TransactionType: valueobject.TransactionType(i.TransactionType),
		Status:          categoryvalueobject.Status(i.Status),
		CreatedAt:       createdAt,
		UpdatedAt:       updatedAt,
	}

	return category, nil
}

func (r *DynamoDBCategoryRepository) Create(ctx context.Context, userID string, category entity.Category) error {
	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":               fmt.Sprintf("USER#%s", userID),
					"SK":               fmt.Sprintf("CATEGORY#%s", category.ID),
					"id":               category.ID,
					"name":             category.Name,
					"transaction_type": string(category.TransactionType),
					"status":           string(category.Status),
					"created_at":       category.CreatedAt.Format(time.RFC3339Nano),
					"updated_at":       category.UpdatedAt.Format(time.RFC3339Nano),
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
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
	}

	err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrCategoryExists
		}

		return fmt.Errorf("create category: %w", err)
	}

	return nil
}

func (r *DynamoDBCategoryRepository) Find(ctx context.Context, userID string, query port.CategoryQuery) (port.CategoryPage, error) {
	conditionExpression := "PK = :pk AND begins_with(SK, :sk)"
	expressionValues := map[string]any{":pk": fmt.Sprintf("USER#%s", userID), ":sk": "CATEGORY#"}

	var filters []string
	expressionNames := map[string]string{}
	if query.TransactionType != nil {
		filters = append(filters, "transaction_type = :transaction_type")
		expressionValues[":transaction_type"] = string(*query.TransactionType)
	}
	if query.Name != nil {
		filters = append(filters, "#name = :name")
		expressionNames["#name"] = "name"
		expressionValues[":name"] = *query.Name
	}
	if query.Status != nil {
		filters = append(filters, "#status = :status")
		expressionNames["#status"] = "status"
		expressionValues[":status"] = string(*query.Status)
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

	queryInput := &dynamodbclient.QueryInput{
		TableName:                 r.tableName,
		Limit:                     aws.Int32(query.Limit),
		ScanIndexForward:          aws.Bool(true),
		KeyConditionExpression:    aws.String(conditionExpression),
		ExpressionAttributeValues: expressionValues,
		ExclusiveStartKey:         startKey,
	}

	if filterExpression != "" {
		queryInput.FilterExpression = aws.String(filterExpression)
	}

	if len(expressionNames) > 0 {
		queryInput.ExpressionAttributeNames = expressionNames
	}

	var queryItems []findCategoryItem
	metadata, err := r.client.Query(ctx, queryInput, &queryItems)
	if err != nil {
		return port.CategoryPage{}, fmt.Errorf("find categories: %w", err)
	}

	categories := make([]entity.Category, 0, len(queryItems))
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

func (r *DynamoDBCategoryRepository) FindOne(ctx context.Context, userID string, categoryID string) (entity.Category, error) {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("CATEGORY#%s", categoryID)}

	var findItem findCategoryItem
	if err := r.client.GetItem(ctx, &dynamodbclient.GetItemInput{TableName: r.tableName, Key: key}, &findItem); err != nil {
		if errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return entity.Category{}, port.ErrCategoryNotFound
		}

		return entity.Category{}, fmt.Errorf("find category: %w", err)
	}

	category, err := findItem.toDomain()
	if err != nil {
		return entity.Category{}, fmt.Errorf("parse category: %w", err)
	}

	return category, nil
}

func (r *DynamoDBCategoryRepository) Update(ctx context.Context, userID string, category entity.Category) error {
	current, err := r.FindOne(ctx, userID, category.ID)
	if err != nil {
		return fmt.Errorf("get current category: %w", err)
	}

	pk := fmt.Sprintf("USER#%s", userID)
	categorySK := fmt.Sprintf("CATEGORY#%s", category.ID)
	oldUpdatedAt := current.UpdatedAt.Format(time.RFC3339Nano)
	updatedAt := category.UpdatedAt.Format(time.RFC3339Nano)
	currentKey := map[string]any{"PK": pk, "SK": categorySK}

	if current.Name == category.Name {
		condition := "updated_at = :old_updated_at"
		expression := "SET #status = :status, updated_at = :updated_at"
		expressionNames := map[string]string{"#status": "status"}
		expressionValues := map[string]any{
			":status":         string(category.Status),
			":updated_at":     updatedAt,
			":old_updated_at": oldUpdatedAt,
		}
		if err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
			TableName:                 r.tableName,
			Key:                       currentKey,
			ConditionExpression:       &condition,
			UpdateExpression:          expression,
			ExpressionAttributeNames:  expressionNames,
			ExpressionAttributeValues: expressionValues,
		}); err != nil {
			if errors.Is(err, dynamodbclient.ErrConditionFailed) {
				return port.ErrConcurrentModification
			}

			return fmt.Errorf("update category: %w", err)
		}
		return nil
	}

	updateCondition := "updated_at = :old_updated_at"
	deleteCondition := "attribute_exists(PK)"
	putCondition := "attribute_not_exists(PK)"
	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Update: &dynamodbclient.TransactUpdate{
				TableName:                r.tableName,
				Key:                      currentKey,
				UpdateExpression:         "SET #name = :name, #status = :status, updated_at = :updated_at",
				ConditionExpression:      &updateCondition,
				ExpressionAttributeNames: map[string]string{"#name": "name", "#status": "status"},
				ExpressionAttributeValues: map[string]any{
					":name":           category.Name,
					":status":         string(category.Status),
					":updated_at":     category.UpdatedAt.Format(time.RFC3339Nano),
					":old_updated_at": oldUpdatedAt,
				},
			},
		},
		{
			Delete: &dynamodbclient.TransactDelete{
				TableName: r.tableName,
				Key: map[string]any{
					"PK": pk,
					"SK": fmt.Sprintf("CATEGORY_NAME#%s", current.Name),
				},
				ConditionExpression: &deleteCondition,
			},
		},
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":          pk,
					"SK":          fmt.Sprintf("CATEGORY_NAME#%s", category.Name),
					"category_id": category.ID,
				},
				ConditionExpression: &putCondition,
			},
		},
	}

	err = r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrCategoryExists
		}
		return fmt.Errorf("update category: %w", err)
	}

	return nil
}

func (r *DynamoDBCategoryRepository) Archive(ctx context.Context, userID string, categoryID string) error {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("CATEGORY#%s", categoryID)}
	expression := "SET #status = :status, updated_at = :updated_at"
	condition := "attribute_exists(PK)"
	expressionNames := map[string]string{"#status": "status"}
	expressionValues := map[string]any{
		":status":     string(categoryvalueobject.StatusArchived),
		":updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:                 r.tableName,
		Key:                       key,
		UpdateExpression:          expression,
		ConditionExpression:       aws.String(condition),
		ExpressionAttributeNames:  expressionNames,
		ExpressionAttributeValues: expressionValues,
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrCategoryNotFound
		}

		return fmt.Errorf("archive category: %w", err)
	}

	return nil
}

func (r *DynamoDBCategoryRepository) Restore(ctx context.Context, userID string, categoryID string) error {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("CATEGORY#%s", categoryID)}
	expression := "SET #status = :status, updated_at = :updated_at"
	condition := "attribute_exists(PK)"
	expressionNames := map[string]string{"#status": "status"}
	expressionValues := map[string]any{
		":status":     string(categoryvalueobject.StatusActive),
		":updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:                 r.tableName,
		Key:                       key,
		UpdateExpression:          expression,
		ConditionExpression:       aws.String(condition),
		ExpressionAttributeNames:  expressionNames,
		ExpressionAttributeValues: expressionValues,
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrCategoryNotFound
		}

		return fmt.Errorf("restore category: %w", err)
	}

	return nil
}
