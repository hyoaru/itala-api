package transaction

import (
	"context"
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

type findTransactionItem struct {
	ID              string                `dynamodbav:"id"`
	Amount          attributevalue.Number `dynamodbav:"amount"`
	TransactionType string                `dynamodbav:"transaction_type"`
	CategoryID      string                `dynamodbav:"category_id"`
	Description     string                `dynamodbav:"description"`
	OccurredAt      string                `dynamodbav:"occurred_at"`
	CreatedAt       string                `dynamodbav:"created_at"`
	UpdatedAt       string                `dynamodbav:"updated_at"`
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
		Type:       valueobjects.TransactionType(i.TransactionType),
		CategoryID: i.CategoryID,
		Amount:     amount,
		OccurredAt: occurredAt,
		CreatedAt:  createdAt,
		UpdatedAt:  updatedAt,
	}

	return transaction, nil
}

func (r *DynamoDBTransactionRepository) Create(ctx context.Context, userID string, transaction entities.Transaction) error {
	return r.client.PutItem(ctx, r.tableName, map[string]any{
		"PK":               fmt.Sprintf("USER#%s", userID),
		"SK":               fmt.Sprintf("TRANSACTION#%s#%s", transaction.OccurredAt.Format(time.RFC3339Nano), transaction.ID),
		"id":               transaction.ID,
		"amount":           dynamodbclient.Decimal(transaction.Amount),
		"transaction_type": string(transaction.Type),
		"category_id":      transaction.CategoryID,
		"description":      transaction.Description,
		"occurred_at":      transaction.OccurredAt.Format(time.RFC3339Nano),
		"created_at":       transaction.CreatedAt.Format(time.RFC3339Nano),
		"updated_at":       transaction.UpdatedAt.Format(time.RFC3339Nano),
	})
}

func (r *DynamoDBTransactionRepository) Find(ctx context.Context, userID string, query port.TransactionQuery) ([]entities.Transaction, error) {
	conditionExpression := "PK = :pk "
	expressionValues := map[string]any{":pk": fmt.Sprintf("USER#%s", userID)}

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
		conditionExpression += " AND begins_with(SK, :sk)"
		expressionValues[":sk"] = "TRANSACTION#"
	}

	// Filters condition
	var filters []string
	if query.Type != nil {
		filters = append(filters, "transaction_type = :type")
		expressionValues[":type"] = string(*query.Type)
	}
	if query.CategoryID != nil {
		filters = append(filters, "category_id = :category_id")
		expressionValues[":category_id"] = *query.CategoryID
	}
	filterExpression := strings.Join(filters, " AND ")

	var queryItems []findTransactionItem
	if err := r.client.Query(ctx, r.tableName, conditionExpression, filterExpression, expressionValues, &queryItems); err != nil {
		return nil, fmt.Errorf("find transactions: %w", err)
	}

	transactions := make([]entities.Transaction, 0, len(queryItems))
	for _, item := range queryItems {
		transaction, err := item.toDomain()
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}

	return transactions, nil
}
