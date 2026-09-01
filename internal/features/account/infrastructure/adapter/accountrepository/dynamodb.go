package account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBAccountRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBAccountRepository {
	return &DynamoDBAccountRepository{client: client, tableName: tableName}
}

type findAccountItem struct {
	ID        string                `dynamodbav:"id"`
	Name      string                `dynamodbav:"name"`
	Balance   attributevalue.Number `dynamodbav:"balance"`
	DeletedAt *string               `dynamodbav:"deleted_at"`
	CreatedAt string                `dynamodbav:"created_at"`
	UpdatedAt string                `dynamodbav:"updated_at"`
}

func (i findAccountItem) toDomain() (entity.Account, error) {
	balance, err := valueobject.NewDecimal(i.Balance.String())
	if err != nil {
		return entity.Account{}, fmt.Errorf("parse balance: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, i.CreatedAt)
	if err != nil {
		return entity.Account{}, fmt.Errorf("parse created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, i.UpdatedAt)
	if err != nil {
		return entity.Account{}, fmt.Errorf("parse updated_at: %w", err)
	}

	var deletedAt *time.Time
	if i.DeletedAt != nil {
		parsed, err := time.Parse(time.RFC3339Nano, *i.DeletedAt)
		if err != nil {
			return entity.Account{}, fmt.Errorf("parse deleted_at: %w", err)
		}
		deletedAt = &parsed
	}

	account := entity.Account{
		ID:        i.ID,
		Name:      i.Name,
		Balance:   balance,
		DeletedAt: deletedAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	return account, nil
}

func (r *DynamoDBAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
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
				ConditionExpression: aws.String("attribute_not_exists(PK)"),
			},
		},
	}

	err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrAccountExists
		}

		return fmt.Errorf("create account: %w", err)
	}

	return nil
}

func (r *DynamoDBAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	conditionExpression := "PK = :pk AND begins_with(SK, :sk)"
	expressionValues := map[string]any{":pk": fmt.Sprintf("USER#%s", userID), ":sk": "ACCOUNT#"}

	var filters []string
	expressionNames := map[string]string{}
	filters = append(filters, "attribute_not_exists(deleted_at)")
	if query.Name != nil {
		filters = append(filters, "#name = :name")
		expressionNames["#name"] = "name"
		expressionValues[":name"] = *query.Name
	}
	filterExpression := strings.Join(filters, " AND ")

	var startKey map[string]any
	if query.Cursor != nil {
		decodedCursor, err := base64.RawURLEncoding.DecodeString(*query.Cursor)
		if err != nil {
			return port.AccountPage{}, fmt.Errorf("decode cursor: %w", err)
		}
		if err := json.Unmarshal(decodedCursor, &startKey); err != nil {
			return port.AccountPage{}, fmt.Errorf("unmarshal start key: %w", err)
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

	var queryItems []findAccountItem
	metadata, err := r.client.Query(ctx, queryInput, &queryItems)
	if err != nil {
		return port.AccountPage{}, fmt.Errorf("find accounts: %w", err)
	}

	accounts := make([]entity.Account, 0, len(queryItems))
	for _, item := range queryItems {
		account, err := item.toDomain()
		if err != nil {
			return port.AccountPage{}, err
		}
		accounts = append(accounts, account)
	}

	if metadata.LastEvaluatedKey == nil {
		return port.AccountPage{Accounts: accounts, NextCursor: nil}, nil
	}

	encodedNextCursor, err := json.Marshal(metadata.LastEvaluatedKey)
	if err != nil {
		return port.AccountPage{}, fmt.Errorf("marshal next cursor: %w", err)
	}

	nextCursor := base64.RawURLEncoding.EncodeToString(encodedNextCursor)
	return port.AccountPage{Accounts: accounts, NextCursor: &nextCursor}, nil
}

func (r *DynamoDBAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("ACCOUNT#%s", id)}

	var findItem findAccountItem
	if err := r.client.GetItem(ctx, &dynamodbclient.GetItemInput{TableName: r.tableName, Key: key}, &findItem); err != nil {
		if errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return entity.Account{}, port.ErrAccountNotFound
		}

		return entity.Account{}, fmt.Errorf("find account: %w", err)
	}

	account, err := findItem.toDomain()
	if err != nil {
		return entity.Account{}, fmt.Errorf("parse account: %w", err)
	}

	if account.DeletedAt != nil {
		return entity.Account{}, port.ErrAccountNotFound
	}

	return account, nil
}

func (r *DynamoDBAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	current, err := r.FindOne(ctx, userID, account.ID)
	if err != nil {
		return fmt.Errorf("get current account: %w", err)
	}

	pk := fmt.Sprintf("USER#%s", userID)
	accountSK := fmt.Sprintf("ACCOUNT#%s", account.ID)
	oldUpdatedAt := current.UpdatedAt.Format(time.RFC3339Nano)
	updatedAt := account.UpdatedAt.Format(time.RFC3339Nano)
	currentKey := map[string]any{"PK": pk, "SK": accountSK}

	if current.Name == account.Name {
		condition := "updated_at = :old_updated_at"
		expression := "SET updated_at = :updated_at"
		expressionValues := map[string]any{
			":updated_at":     updatedAt,
			":old_updated_at": oldUpdatedAt,
		}

		if err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
			TableName:                 r.tableName,
			Key:                       currentKey,
			ConditionExpression:       &condition,
			UpdateExpression:          expression,
			ExpressionAttributeValues: expressionValues,
		}); err != nil {
			if errors.Is(err, dynamodbclient.ErrConditionFailed) {
				return port.ErrConcurrentModification
			}

			return fmt.Errorf("update account: %w", err)
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
				UpdateExpression:         "SET #name = :name, updated_at = :updated_at",
				ConditionExpression:      &updateCondition,
				ExpressionAttributeNames: map[string]string{"#name": "name"},
				ExpressionAttributeValues: map[string]any{
					":name":           account.Name,
					":updated_at":     updatedAt,
					":old_updated_at": oldUpdatedAt,
				},
			},
		},
		{
			Delete: &dynamodbclient.TransactDelete{
				TableName: r.tableName,
				Key: map[string]any{
					"PK": pk,
					"SK": fmt.Sprintf("ACCOUNT_NAME#%s", current.Name),
				},
				ConditionExpression: &deleteCondition,
			},
		},
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         pk,
					"SK":         fmt.Sprintf("ACCOUNT_NAME#%s", account.Name),
					"account_id": account.ID,
				},
				ConditionExpression: &putCondition,
			},
		},
	}

	err = r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrAccountExists
		}

		return fmt.Errorf("update account: %w", err)
	}

	return nil
}

func (r *DynamoDBAccountRepository) Delete(ctx context.Context, userID string, id string) error {
	current, err := r.FindOne(ctx, userID, id)
	if err != nil {
		return err
	}

	pk := fmt.Sprintf("USER#%s", userID)
	now := time.Now().UTC().Format(time.RFC3339Nano)

	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Update: &dynamodbclient.TransactUpdate{
				TableName: r.tableName,
				Key: map[string]any{
					"PK": pk,
					"SK": fmt.Sprintf("ACCOUNT#%s", id),
				},
				UpdateExpression:    "SET deleted_at = :deleted_at, updated_at = :updated_at",
				ConditionExpression: aws.String("attribute_exists(PK)"),
				ExpressionAttributeValues: map[string]any{
					":deleted_at": now,
					":updated_at": now,
				},
			},
		},
		{
			Delete: &dynamodbclient.TransactDelete{
				TableName: r.tableName,
				Key: map[string]any{
					"PK": pk,
					"SK": fmt.Sprintf("ACCOUNT_NAME#%s", current.Name),
				},
			},
		},
	}

	err = r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrAccountNotFound
		}

		return fmt.Errorf("delete account: %w", err)
	}

	return nil
}

func (r *DynamoDBAccountRepository) AdjustBalance(ctx context.Context, userID string, accountID string, idempotencyKey string, delta valueobject.Decimal) error {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("ACCOUNT#%s", accountID)}
	expression := "SET balance = balance + :delta, updated_at = :updated_at"
	condition := "attribute_exists(PK)"
	expressionValues := map[string]any{
		":delta":      dynamodbclient.Decimal(delta),
		":updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:                 r.tableName,
		Key:                       key,
		UpdateExpression:          expression,
		ConditionExpression:       aws.String(condition),
		ExpressionAttributeValues: expressionValues,
	})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrAccountNotFound
		}

		return fmt.Errorf("adjust account balance: %w", err)
	}

	return nil
}
