package account

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	accountvalueobject "github.com/hyoaru/itala-api/internal/features/account/domain/valueobject"
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
	Status    string                `dynamodbav:"status"`
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

	account := entity.Account{
		ID:        i.ID,
		Name:      i.Name,
		Balance:   balance,
		Status:    accountvalueobject.Status(i.Status),
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
					"status":     string(account.Status),
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
				ConditionExpression: "attribute_not_exists(PK)",
			},
		},
	}

	err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrItemExists) {
			return port.ErrAccountExists
		}
	}

	return err
}

func (r *DynamoDBAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	conditionExpression := "PK = :pk AND begins_with(SK, :sk)"
	expressionValues := map[string]any{":pk": fmt.Sprintf("USER#%s", userID), ":sk": "ACCOUNT#"}

	var filters []string
	expressionNames := map[string]string{}
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
			return port.AccountPage{}, fmt.Errorf("decode cursor: %w", err)
		}
		if err := json.Unmarshal(decodedCursor, &startKey); err != nil {
			return port.AccountPage{}, fmt.Errorf("unmarshal start key: %w", err)
		}
	}

	var queryItems []findAccountItem
	metadata, err := r.client.Query(ctx, &dynamodbclient.QueryInput{
		TableName:                 r.tableName,
		Limit:                     query.Limit,
		ScanIndexForward:          true,
		KeyConditionExpression:    conditionExpression,
		FilterExpression:          filterExpression,
		ExpressionAttributeNames:  expressionNames,
		ExpressionAttributeValues: expressionValues,
		ExclusiveStartKey:         startKey,
		Output:                    &queryItems,
	})
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
	if err := r.client.GetItem(ctx, &dynamodbclient.GetItemInput{TableName: r.tableName, Key: key, Output: &findItem}); err != nil {
		if errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return entity.Account{}, port.ErrAccountNotFound
		}

		return entity.Account{}, fmt.Errorf("find account: %w", err)
	}

	account, err := findItem.toDomain()
	if err != nil {
		return entity.Account{}, fmt.Errorf("parse account: %w", err)
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
	updatedAt := account.UpdatedAt.Format(time.RFC3339Nano)
	currentKey := map[string]any{"PK": pk, "SK": accountSK}

	if current.Name == account.Name {
		expression := "SET #status = :status, updated_at = :updated_at"
		expressionNames := map[string]string{"#status": "status"}
		expressionValues := map[string]any{":status": string(account.Status), ":updated_at": updatedAt}
		if err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
			TableName:                 r.tableName,
			Key:                       currentKey,
			UpdateExpression:          expression,
			ExpressionAttributeNames:  expressionNames,
			ExpressionAttributeValues: expressionValues,
		}); err != nil {
			return err
		}
		return nil
	}

	transactItems := []dynamodbclient.TransactWriteItem{
		{
			Update: &dynamodbclient.TransactUpdate{
				TableName:                r.tableName,
				Key:                      currentKey,
				UpdateExpression:         "SET #name = :name, #status = :status, updated_at = :updated_at",
				ConditionExpression:      "#name = :old_name",
				ExpressionAttributeNames: map[string]string{"#name": "name", "#status": "status"},
				ExpressionAttributeValues: map[string]any{
					":name":       account.Name,
					":status":     string(account.Status),
					":updated_at": updatedAt,
					":old_name":   current.Name,
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
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":         pk,
					"SK":         fmt.Sprintf("ACCOUNT_NAME#%s", account.Name),
					"account_id": account.ID,
				},
				ConditionExpression: "attribute_not_exists(PK)",
			},
		},
	}

	err = r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: transactItems})
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrItemExists) {
			return port.ErrAccountExists
		}
	}

	return err
}

func (r *DynamoDBAccountRepository) Archive(ctx context.Context, userID string, id string) error {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("ACCOUNT#%s", id)}
	expression := "SET #status = :status, updated_at = :updated_at"
	condition := "#status = :active"
	expressionNames := map[string]string{"#status": "status"}
	expressionValues := map[string]any{
		":status":     string(accountvalueobject.StatusArchived),
		":active":     string(accountvalueobject.StatusActive),
		":updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:                 r.tableName,
		Key:                       key,
		UpdateExpression:          expression,
		ConditionExpression:       condition,
		ExpressionAttributeNames:  expressionNames,
		ExpressionAttributeValues: expressionValues,
	})
	if err == nil {
		return nil
	}

	if !errors.Is(err, dynamodbclient.ErrConditionFailed) {
		return err
	}

	current, err := r.FindOne(ctx, userID, id)
	if err != nil {
		return fmt.Errorf("get current account: %w", err)
	}

	if current.Status == accountvalueobject.StatusArchived {
		return nil
	}

	return fmt.Errorf("archive account: unexpected status %q", current.Status)
}

func (r *DynamoDBAccountRepository) Restore(ctx context.Context, userID string, id string) error {
	key := map[string]any{"PK": fmt.Sprintf("USER#%s", userID), "SK": fmt.Sprintf("ACCOUNT#%s", id)}
	expression := "SET #status = :status, updated_at = :updated_at"
	condition := "#status = :archived"
	expressionNames := map[string]string{"#status": "status"}
	expressionValues := map[string]any{
		":status":     string(accountvalueobject.StatusActive),
		":archived":   string(accountvalueobject.StatusArchived),
		":updated_at": time.Now().UTC().Format(time.RFC3339Nano),
	}

	err := r.client.UpdateItem(ctx, &dynamodbclient.UpdateItemInput{
		TableName:                 r.tableName,
		Key:                       key,
		UpdateExpression:          expression,
		ConditionExpression:       condition,
		ExpressionAttributeNames:  expressionNames,
		ExpressionAttributeValues: expressionValues,
	})
	if err == nil {
		return nil
	}

	if !errors.Is(err, dynamodbclient.ErrConditionFailed) {
		return err
	}

	current, err := r.FindOne(ctx, userID, id)
	if err != nil {
		return fmt.Errorf("get current account: %w", err)
	}

	if current.Status == accountvalueobject.StatusActive {
		return nil
	}

	return fmt.Errorf("restore account: unexpected status %q", current.Status)
}
