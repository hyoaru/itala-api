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

func (r *DynamoDBAccountRepository) Find(ctx context.Context, userID string, query port.AccountQuery) (port.AccountPage, error) {
	conditionExpression := "PK = :pk AND begins_with(SK, :sk)"
	expressionValues := map[string]any{":pk": fmt.Sprintf("USER#%s", userID), ":sk": "ACCOUNT#"}

	var filters []string
	if query.Name != nil {
		filters = append(filters, "name = :name")
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

	var queryItems []findAccountItem
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
