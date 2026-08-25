package transaction

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	port "github.com/hyoaru/itala-api/internal/features/transaction/application/ports/transactionrepository"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBTransactionRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBTransactionRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBTransactionRepository {
	return &DynamoDBTransactionRepository{client: client, tableName: tableName}
}

type transactionIndex struct {
	PK string
	SK string
}

type findTransactionItem struct {
	ID          string                `dynamodbav:"id"`
	Amount      attributevalue.Number `dynamodbav:"amount"`
	Type        string                `dynamodbav:"type"`
	AccountID   string                `dynamodbav:"account_id"`
	CategoryID  string                `dynamodbav:"category_id"`
	Description string                `dynamodbav:"description"`
	OccurredAt  string                `dynamodbav:"occurred_at"`
	CreatedAt   string                `dynamodbav:"created_at"`
	UpdatedAt   string                `dynamodbav:"updated_at"`
}

func (i findTransactionItem) toDomain() (entities.Transaction, error) {
	amount, err := valueobjects.NewDecimal(i.Amount.String())
	if err != nil {
		return entities.Transaction{}, fmt.Errorf("parse amount: %w", err)
	}

	occurredAt, err := time.Parse(time.RFC3339Nano, i.OccurredAt)
	if err != nil {
		return entities.Transaction{}, fmt.Errorf("parse occurred_at: %w", err)
	}

	createdAt, err := time.Parse(time.RFC3339Nano, i.CreatedAt)
	if err != nil {
		return entities.Transaction{}, fmt.Errorf("parse created_at: %w", err)
	}

	updatedAt, err := time.Parse(time.RFC3339Nano, i.UpdatedAt)
	if err != nil {
		return entities.Transaction{}, fmt.Errorf("parse updated_at: %w", err)
	}

	transaction := entities.Transaction{
		ID:         i.ID,
		Amount:     amount,
		Type:       valueobjects.TransactionType(i.Type),
		CategoryID: i.CategoryID,
		AccountID:  i.AccountID,
		OccurredAt: occurredAt,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	return transaction, nil
}

func (r *DynamoDBTransactionRepository) Create(ctx context.Context, userID string, transaction entities.Transaction) error {
	occurredAt := transaction.OccurredAt.Format(time.RFC3339Nano)
	sortKey := fmt.Sprintf("TRANSACTION#%s%s", occurredAt, transaction.ID)

	return r.client.PutItem(ctx, r.tableName, map[string]any{
		// Base table
		"PK": fmt.Sprintf("USER#%s", userID),
		"SK": sortKey,

		// Type GSI
		"GSI1PK": fmt.Sprintf("USER#%s#TYPE#%s", userID, string(transaction.Type)),
		"GSI1SK": sortKey,

		// Account GSI
		"GSI2PK": fmt.Sprintf("USER#%s#ACCOUNT#%s", userID, transaction.AccountID),
		"GSI2SK": sortKey,

		// Category GSI
		"GSI3PK": fmt.Sprintf("USER#%s#CATEGORY#%s", userID, transaction.CategoryID),
		"GSI3SK": sortKey,

		"id":          transaction.ID,
		"amount":      dynamodbclient.Decimal(transaction.Amount),
		"type":        string(transaction.Type),
		"account_id":  transaction.AccountID,
		"category_id": transaction.CategoryID,
		"description": transaction.Description,
		"occurred_at": transaction.OccurredAt.Format(time.RFC3339Nano),
		"created_at":  transaction.CreatedAt.Format(time.RFC3339Nano),
		"updated_at":  transaction.UpdatedAt.Format(time.RFC3339Nano),
	})
}

func (r *DynamoDBTransactionRepository) findByIndex(ctx context.Context, index transactionIndex, pk string, query port.TransactionQuery) (port.TransactionPage, error) {
	conditionExpression := fmt.Sprintf("%s = :pk ", index.PK)
	expressionValues := map[string]any{":pk": pk}

	// Sort key condition
	switch {
	case query.From != nil && query.To != nil:
		conditionExpression += " AND SK BETWEEN :from AND :to"
		expressionValues[":from"] = fmt.Sprintf("TRANSACTION#%s", query.From.UTC().Format(time.RFC3339Nano))
		expressionValues[":to"] = fmt.Sprintf("TRANSACTION#%s~", query.To.UTC().Format(time.RFC3339Nano))
	case query.From != nil:
		conditionExpression += " AND SK >= :from"
		expressionValues[":from"] = fmt.Sprintf("TRANSACTION#%s", query.From.UTC().Format(time.RFC3339Nano))
	case query.To != nil:
		conditionExpression += " AND SK <= :to"
		expressionValues[":to"] = fmt.Sprintf("TRANSACTION#%s~", query.To.UTC().Format(time.RFC3339Nano))
	default:
		conditionExpression += fmt.Sprintf(" AND begins_with(%s, :sk)", index.SK)
		expressionValues[":sk"] = "TRANSACTION#"
	}

	var filters []string
	if query.Type != nil {
		filters = append(filters, "type = :type")
		expressionValues[":type"] = string(*query.Type)
	}
	if query.CategoryID != nil {
		filters = append(filters, "category_id = :category_id")
		expressionValues[":category_id"] = *query.CategoryID
	}
	if query.AccountID != nil {
		filters = append(filters, "account_id = :account_id")
		expressionValues[":account_id"] = *query.AccountID
	}
	filterExpression := strings.Join(filters, " AND ")

	var startKey map[string]any
	if query.Cursor != nil {
		decodedCursor, err := base64.RawURLEncoding.DecodeString(*query.Cursor)
		if err != nil {
			return port.TransactionPage{}, fmt.Errorf("decode cursor: %w", err)
		}
		if err := json.Unmarshal(decodedCursor, &startKey); err != nil {
			return port.TransactionPage{}, fmt.Errorf("unmarshal start key: %w", err)
		}
	}

	var queryItems []findTransactionItem
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
		return port.TransactionPage{}, fmt.Errorf("find transactions: %w", err)
	}

	transactions := make([]entities.Transaction, 0, len(queryItems))
	for _, item := range queryItems {
		transaction, err := item.toDomain()
		if err != nil {
			return port.TransactionPage{}, err
		}
		transactions = append(transactions, transaction)
	}

	if metadata.LastEvaluatedKey == nil {
		return port.TransactionPage{Transactions: transactions, NextCursor: nil}, nil
	}

	encodedNextCursor, err := json.Marshal(metadata.LastEvaluatedKey)
	if err != nil {
		return port.TransactionPage{}, fmt.Errorf("marshal next cursor: %w", err)
	}

	nextCursor := base64.RawURLEncoding.EncodeToString(encodedNextCursor)
	return port.TransactionPage{Transactions: transactions, NextCursor: &nextCursor}, nil
}

func (r *DynamoDBTransactionRepository) Find(ctx context.Context, userID string, query port.TransactionQuery) (port.TransactionPage, error) {
	switch {
	case query.Type != nil:
		return r.findByIndex(ctx, transactionIndex{PK: "GSI1PK", SK: "GSI1SK"}, fmt.Sprintf("USER#%s#TYPE#%s", userID, *query.Type), query)
	case query.AccountID != nil:
		return r.findByIndex(ctx, transactionIndex{PK: "GSI2PK", SK: "GSI2SK"}, fmt.Sprintf("USER#%s#ACCOUNT#%s", userID, *query.AccountID), query)
	case query.CategoryID != nil:
		return r.findByIndex(ctx, transactionIndex{PK: "GSI3PK", SK: "GSI3SK"}, fmt.Sprintf("USER#%s#CATEGORY#%s", userID, *query.CategoryID), query)
	default:
		return r.findByIndex(ctx, transactionIndex{PK: "PK", SK: "SK"}, fmt.Sprintf("USER#%s", userID), query)
	}
}
