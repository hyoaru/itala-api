package ledgerrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"

	port "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type DynamoDBLedgerRepository struct {
	client    dynamodbclient.DynamoDBClient
	tableName string
}

func NewDynamoDBLedgerRepository(client dynamodbclient.DynamoDBClient, tableName string) *DynamoDBLedgerRepository {
	return &DynamoDBLedgerRepository{client: client, tableName: tableName}
}

type ledgerItem struct {
	SK                      string  `dynamodbav:"SK"`
	ID                      string  `dynamodbav:"id"`
	TransactionID           string  `dynamodbav:"transaction_id"`
	AccountID               string  `dynamodbav:"account_id"`
	Description             string  `dynamodbav:"description"`
	OccurredAt              string  `dynamodbav:"occurred_at"`
	CreatedAt               string  `dynamodbav:"created_at"`
	Amount                  string  `dynamodbav:"amount"`
	ReversesTransactionID   *string `dynamodbav:"reverses_transaction_id"`
	ReversedByTransactionID *string `dynamodbav:"reversed_by_transaction_id"`
}

type idempotencyItem struct {
	TransactionID string `dynamodbav:"transaction_id"`
	RequestHash   string `dynamodbav:"request_hash"`
	CreatedAt     string `dynamodbav:"created_at"`
}

func (r *DynamoDBLedgerRepository) RecordTransaction(ctx context.Context, userID string, input port.RecordTransactionInput) error {
	userPK := fmt.Sprintf("USER#%s", userID)
	condition := "attribute_not_exists(PK)"

	items := []dynamodbclient.TransactWriteItem{
		{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":                         userPK,
					"SK":                         fmt.Sprintf("TRANSACTION#%s#%s", input.OccurredAt, input.TransactionID),
					"id":                         input.TransactionID,
					"description":                input.Description,
					"occurred_at":                input.OccurredAt,
					"created_at":                 input.CreatedAt,
					"reverses_transaction_id":    nil,
					"reversed_by_transaction_id": nil,
				},
				ConditionExpression: &condition,
			},
		},
	}

	for _, posting := range input.Postings {
		items = append(items, dynamodbclient.TransactWriteItem{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":             userPK,
					"SK":             fmt.Sprintf("TRANSACTION#%s#%s#POSTING#%s", input.OccurredAt, input.TransactionID, posting.ID),
					"id":             posting.ID,
					"transaction_id": input.TransactionID,
					"account_id":     posting.AccountID,
					"amount":         posting.Amount.String(),
				},
				ConditionExpression: &condition,
			},
		})
	}

	items = append(items, dynamodbclient.TransactWriteItem{
		Put: &dynamodbclient.TransactPut{
			TableName: r.tableName,
			Item: map[string]any{
				"PK":                      userPK,
				"SK":                      fmt.Sprintf("TRANSACTION_SUMMARY#%s#%s", input.OccurredAt, input.TransactionID),
				"id":                      input.Summary.ID,
				"transaction_id":          input.Summary.TransactionID,
				"description":             input.Summary.Description,
				"occurred_at":             input.Summary.OccurredAt,
				"display_amount":          input.Summary.DisplayAmount.String(),
				"display_account_id":      input.Summary.DisplayAccountID,
				"counterparty_account_id": input.Summary.CounterpartyAccountID,
			},
			ConditionExpression: &condition,
		},
	})

	for _, entry := range input.StatementEntries {
		items = append(items, dynamodbclient.TransactWriteItem{
			Put: &dynamodbclient.TransactPut{
				TableName: r.tableName,
				Item: map[string]any{
					"PK":                      userPK,
					"SK":                      fmt.Sprintf("ACCOUNT_STATEMENT_ENTRY#%s#%s#%s", entry.AccountID, entry.TransactionOccurredAt, entry.ID),
					"id":                      entry.ID,
					"transaction_id":          entry.TransactionID,
					"account_id":              entry.AccountID,
					"amount":                  entry.Amount.String(),
					"description":             entry.Description,
					"transaction_occurred_at": entry.TransactionOccurredAt,
					"counterparty_account_id": entry.CounterpartyAccountID,
				},
				ConditionExpression: &condition,
			},
		})
	}

	for _, delta := range input.BalanceDeltas {
		updateExpression := "SET account_id = if_not_exists(account_id, :account_id), updated_at = if_not_exists(updated_at, :updated_at) ADD #balance :delta"
		items = append(items, dynamodbclient.TransactWriteItem{
			Update: &dynamodbclient.TransactUpdate{
				TableName:        r.tableName,
				Key:              map[string]any{"PK": userPK, "SK": fmt.Sprintf("BALANCE#%s", delta.AccountID)},
				UpdateExpression: updateExpression,
				ExpressionAttributeNames: map[string]string{
					"#balance": "balance",
				},
				ExpressionAttributeValues: map[string]any{
					":account_id": delta.AccountID,
					":updated_at": input.CreatedAt,
					":delta":      delta.Amount.String(),
				},
			},
		})
	}

	items = append(items, dynamodbclient.TransactWriteItem{
		Put: &dynamodbclient.TransactPut{
			TableName: r.tableName,
			Item: map[string]any{
				"PK":             fmt.Sprintf("IDEMPOTENCY#%s", input.Idempotency.KeyHash),
				"SK":             "#RESULT",
				"transaction_id": input.Idempotency.TransactionID,
				"request_hash":   input.Idempotency.RequestHash,
				"created_at":     input.Idempotency.CreatedAt,
			},
			ConditionExpression: &condition,
		},
	})

	if err := r.client.TransactWriteItems(ctx, &dynamodbclient.TransactWriteItemsInput{TransactItems: items}); err != nil {
		if errors.Is(err, dynamodbclient.ErrConditionFailed) {
			return port.ErrConditionFailed
		}
		return fmt.Errorf("record transaction: %w", err)
	}

	return nil
}

func (r *DynamoDBLedgerRepository) FindIdempotency(ctx context.Context, userID string, keyHash string) (port.IdempotencyRecord, bool, error) {
	var item idempotencyItem
	err := r.client.GetItem(ctx, &dynamodbclient.GetItemInput{
		TableName: r.tableName,
		Key: map[string]any{
			"PK": fmt.Sprintf("IDEMPOTENCY#%s", keyHash),
			"SK": "#RESULT",
		},
	}, &item)
	if err != nil {
		if errors.Is(err, dynamodbclient.ErrItemNotFound) {
			return port.IdempotencyRecord{}, false, nil
		}
		return port.IdempotencyRecord{}, false, fmt.Errorf("find idempotency: %w", err)
	}

	return port.IdempotencyRecord{
		TransactionID: item.TransactionID,
		RequestHash:   item.RequestHash,
		CreatedAt:     item.CreatedAt,
	}, true, nil
}

func (r *DynamoDBLedgerRepository) FindTransactionDetail(ctx context.Context, userID string, occurredAt string, id string) (entity.Transaction, []entity.Posting, error) {
	prefix := fmt.Sprintf("TRANSACTION#%s#%s", occurredAt, id)
	condition := "PK = :pk AND begins_with(SK, :prefix)"
	consistentRead := true

	var items []ledgerItem
	_, err := r.client.Query(ctx, &dynamodbclient.QueryInput{
		TableName:              r.tableName,
		ScanIndexForward:       aws.Bool(true),
		ConsistentRead:         &consistentRead,
		KeyConditionExpression: &condition,
		ExpressionAttributeValues: map[string]any{
			":pk":     fmt.Sprintf("USER#%s", userID),
			":prefix": prefix,
		},
	}, &items)
	if err != nil {
		return entity.Transaction{}, nil, fmt.Errorf("find transaction detail: %w", err)
	}

	var transaction entity.Transaction
	found := false
	postings := make([]entity.Posting, 0, len(items))

	for _, item := range items {
		switch {
		case item.SK == prefix:
			transaction = entity.Transaction{
				ID:                      item.ID,
				Description:             item.Description,
				OccurredAt:              item.OccurredAt,
				CreatedAt:               item.CreatedAt,
				ReversesTransactionID:   item.ReversesTransactionID,
				ReversedByTransactionID: item.ReversedByTransactionID,
			}
			found = true
		case strings.Contains(item.SK, "#POSTING#"):
			amount, err := valueobject.NewDecimal(item.Amount)
			if err != nil {
				return entity.Transaction{}, nil, fmt.Errorf("parse posting amount: %w", err)
			}
			postings = append(postings, entity.Posting{
				ID:            item.ID,
				TransactionID: item.TransactionID,
				AccountID:     item.AccountID,
				Amount:        amount,
			})
		}
	}

	if !found {
		return entity.Transaction{}, nil, port.ErrTransactionNotFound
	}

	return transaction, postings, nil
}
