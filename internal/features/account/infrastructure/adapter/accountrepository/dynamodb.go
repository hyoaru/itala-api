package accountrepository

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBAccountRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBAccountRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBAccountRepository {
	return &DynamoDBAccountRepository{client: client, tableName: tableName}
}

type accountItem struct {
	ID        string  `dynamodbav:"id"`
	Name      string  `dynamodbav:"name"`
	Type      string  `dynamodbav:"type"`
	DeletedAt *string `dynamodbav:"deleted_at"`
	CreatedAt string  `dynamodbav:"created_at"`
	UpdatedAt string  `dynamodbav:"updated_at"`
}

func parseTime(value string, field string) (time.Time, error) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse %s: %w", field, err)
	}
	return parsed, nil
}

func (i accountItem) toDomain() (entity.Account, error) {
	createdAt, err := parseTime(i.CreatedAt, "created_at")
	if err != nil {
		return entity.Account{}, err
	}

	updatedAt, err := parseTime(i.UpdatedAt, "updated_at")
	if err != nil {
		return entity.Account{}, err
	}

	var deletedAt *time.Time
	if i.DeletedAt != nil {
		parsed, err := parseTime(*i.DeletedAt, "deleted_at")
		if err != nil {
			return entity.Account{}, err
		}
		deletedAt = &parsed
	}

	return entity.Account{
		ID:        i.ID,
		Name:      i.Name,
		Type:      entity.Type(i.Type),
		DeletedAt: deletedAt,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}, nil
}

func (r *DynamoDBAccountRepository) Create(ctx context.Context, userID string, account entity.Account) error {
	normalized := entity.NormalizeName(account.Name)
	userPK := fmt.Sprintf("USER#%s", userID)
	condition := "attribute_not_exists(PK)"

	items := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         userPK,
					"SK":         fmt.Sprintf("ACCOUNT#%s", account.ID),
					"id":         account.ID,
					"name":       account.Name,
					"type":       string(account.Type),
					"created_at": account.CreatedAt.Format(time.RFC3339Nano),
					"updated_at": account.UpdatedAt.Format(time.RFC3339Nano),
				},
				ConditionExpression: &condition,
			},
		},
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         userPK,
					"SK":         fmt.Sprintf("ACCOUNT_NAME#%s", normalized),
					"account_id": account.ID,
				},
				ConditionExpression: &condition,
			},
		},
	}

	if err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: items}); err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrAccountExists
		}
		return fmt.Errorf("create account: %w", err)
	}

	return nil
}

func (r *DynamoDBAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	filter := "attribute_not_exists(deleted_at)"
	names := map[string]string{}
	values := map[string]any{
		":pk":     fmt.Sprintf("USER#%s", userID),
		":prefix": "ACCOUNT#",
	}

	if query.Name != nil {
		filter += " AND #name = :name"
		names["#name"] = "name"
		values[":name"] = *query.Name
	}

	var startKey map[string]any
	if query.Cursor != nil {
		decoded, err := base64.RawURLEncoding.DecodeString(*query.Cursor)
		if err != nil {
			return port.AccountPage{}, fmt.Errorf("decode cursor: %w", err)
		}
		if err := json.Unmarshal(decoded, &startKey); err != nil {
			return port.AccountPage{}, fmt.Errorf("unmarshal cursor: %w", err)
		}
	}

	condition := "PK = :pk AND begins_with(SK, :prefix)"
	input := &dynamodbclient.QueryInput{
		TableName:                 r.tableName,
		Limit:                     aws.Int32(query.Limit),
		ScanIndexForward:          aws.Bool(true),
		KeyConditionExpression:    &condition,
		FilterExpression:          &filter,
		ExpressionAttributeNames:  names,
		ExpressionAttributeValues: values,
		ExclusiveStartKey:         startKey,
	}

	var items []accountItem
	metadata, err := r.client.Query(ctx, input, &items)
	if err != nil {
		return port.AccountPage{}, fmt.Errorf("find accounts: %w", err)
	}

	accounts := make([]entity.Account, 0, len(items))
	for _, item := range items {
		account, err := item.toDomain()
		if err != nil {
			return port.AccountPage{}, err
		}
		accounts = append(accounts, account)
	}

	if metadata.LastEvaluatedKey == nil {
		return port.AccountPage{Accounts: accounts}, nil
	}

	encoded, err := json.Marshal(metadata.LastEvaluatedKey)
	if err != nil {
		return port.AccountPage{}, fmt.Errorf("marshal cursor: %w", err)
	}
	next := base64.RawURLEncoding.EncodeToString(encoded)
	return port.AccountPage{Accounts: accounts, NextCursor: &next}, nil
}

func (r *DynamoDBAccountRepository) FindOne(ctx context.Context, userID string, id string) (entity.Account, error) {
	var item accountItem
	if err := r.client.GetItem(ctx, &dynamodbclient.GetItemInput{
		TableName: r.tableName,
		Key: map[string]any{
			"PK": fmt.Sprintf("USER#%s", userID),
			"SK": fmt.Sprintf("ACCOUNT#%s", id),
		},
	}, &item); err != nil {
		if errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return entity.Account{}, port.ErrAccountNotFound
		}
		return entity.Account{}, fmt.Errorf("find account: %w", err)
	}

	account, err := item.toDomain()
	if err != nil {
		return entity.Account{}, err
	}
	if account.DeletedAt != nil {
		return entity.Account{}, port.ErrAccountNotFound
	}

	return account, nil
}

func (r *DynamoDBAccountRepository) Update(ctx context.Context, userID string, account entity.Account) error {
	current, err := r.FindOne(ctx, userID, account.ID)
	if err != nil {
		return err
	}

	userPK := fmt.Sprintf("USER#%s", userID)
	oldNormalized := entity.NormalizeName(current.Name)
	newNormalized := entity.NormalizeName(account.Name)
	condition := "updated_at = :expected AND attribute_not_exists(deleted_at)"
	lockCondition := "attribute_not_exists(PK)"
	lockExistsCondition := "attribute_exists(PK)"

	items := []dynamodbclient.TransactWriteItem{
		{
			Update: &dynamodbclient.TransactUpdate{
				TableName:           r.tableName,
				Key:                 map[string]any{"PK": userPK, "SK": fmt.Sprintf("ACCOUNT#%s", account.ID)},
				UpdateExpression:    "SET #name = :name, updated_at = :updated_at",
				ConditionExpression: &condition,
				ExpressionAttributeNames: map[string]string{
					"#name": "name",
				},
				ExpressionAttributeValues: map[string]any{
					":name":       account.Name,
					":expected":   current.UpdatedAt.Format(time.RFC3339Nano),
					":updated_at": account.UpdatedAt.Format(time.RFC3339Nano),
				},
			},
		},
	}

	if oldNormalized != newNormalized {
		items = append(items,
			dynamodbclient.TransactWriteItem{
				Delete: &dynamodbclient.TransactDelete{
					TableName:           r.tableName,
					Key:                 map[string]any{"PK": userPK, "SK": fmt.Sprintf("ACCOUNT_NAME#%s", oldNormalized)},
					ConditionExpression: &lockExistsCondition,
				},
			},
			dynamodbclient.TransactWriteItem{
				Put: &dynamodbclient.TransactPut{
					TableName: r.tableName,
					Item: map[string]any{
						"PK":         userPK,
						"SK":         fmt.Sprintf("ACCOUNT_NAME#%s", newNormalized),
						"account_id": account.ID,
					},
					ConditionExpression: &lockCondition,
				},
			},
		)
	}

	if err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: items}); err != nil {
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

	userPK := fmt.Sprintf("USER#%s", userID)
	normalized := entity.NormalizeName(current.Name)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	condition := "attribute_not_exists(deleted_at) AND updated_at = :expected"
	lockExistsCondition := "attribute_exists(PK)"

	items := []dynamodbclient.TransactWriteItem{
		{
			Update: &dynamodbclient.TransactUpdate{
				TableName:           r.tableName,
				Key:                 map[string]any{"PK": userPK, "SK": fmt.Sprintf("ACCOUNT#%s", id)},
				UpdateExpression:    "SET deleted_at = :deleted_at, updated_at = :updated_at",
				ConditionExpression: &condition,
				ExpressionAttributeValues: map[string]any{
					":deleted_at": now,
					":updated_at": now,
					":expected":   current.UpdatedAt.Format(time.RFC3339Nano),
				},
			},
		},
		{
			Delete: &dynamodbclient.TransactDelete{
				TableName:           r.tableName,
				Key:                 map[string]any{"PK": userPK, "SK": fmt.Sprintf("ACCOUNT_NAME#%s", normalized)},
				ConditionExpression: &lockExistsCondition,
			},
		},
	}

	if err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: items}); err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrConcurrentModification
		}
		return fmt.Errorf("delete account: %w", err)
	}

	return nil
}
