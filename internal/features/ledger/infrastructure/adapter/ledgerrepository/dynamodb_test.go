package ledgerrepository

import (
	"context"
	"strings"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type fakeClient struct {
	transactInput *dynamodbclient.TransactWriteItemsInput
	queryInput    *dynamodbclient.QueryInput
	queryItems    []ledgerItem
	getItem       idempotencyItem
}

func (f *fakeClient) PutItem(context.Context, *dynamodbclient.PutItemInput) error { return nil }

func (f *fakeClient) TransactWriteItems(_ context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
	f.transactInput = input
	return nil
}

func (f *fakeClient) GetItem(_ context.Context, _ *dynamodbclient.GetItemInput, output any) error {
	item := output.(*idempotencyItem)
	*item = f.getItem
	return nil
}

func (f *fakeClient) Query(_ context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
	f.queryInput = input
	items := output.(*[]ledgerItem)
	*items = f.queryItems
	return dynamodbclient.QueryOutput{}, nil
}

func (f *fakeClient) UpdateItem(context.Context, *dynamodbclient.UpdateItemInput) error { return nil }
func (f *fakeClient) DeleteItem(context.Context, *dynamodbclient.DeleteItemInput) error { return nil }

func testDecimal(t *testing.T, value string) valueobject.Decimal {
	t.Helper()
	amount, err := valueobject.NewDecimal(value)
	if err != nil {
		t.Fatalf("decimal: %v", err)
	}
	return amount
}

func TestRecordTransactionWritesEveryProjection(t *testing.T) {
	client := &fakeClient{}
	repo := NewDynamoDBLedgerRepository(client, "itala")
	amount := testDecimal(t, "100")
	now := time.Now().UTC().Format(time.RFC3339Nano)

	input := port.RecordTransactionInput{
		TransactionID: "txn-1",
		Description:   "Groceries",
		OccurredAt:    "2026-11-22",
		CreatedAt:     now,
		Postings: []entity.Posting{
			{ID: "p1", TransactionID: "txn-1", AccountID: "cash", Amount: testDecimal(t, "-100")},
			{ID: "p2", TransactionID: "txn-1", AccountID: "groceries", Amount: amount},
		},
		Summary: entity.TransactionSummary{
			ID:                    "txn-1",
			TransactionID:         "txn-1",
			Description:           "Groceries",
			OccurredAt:            "2026-11-22",
			DisplayAmount:         amount,
			DisplayAccountID:      "groceries",
			CounterpartyAccountID: "cash",
		},
		StatementEntries: []entity.AccountStatementEntry{
			{ID: "e1", TransactionID: "txn-1", AccountID: "cash", Amount: testDecimal(t, "-100"), Description: "Groceries", TransactionOccurredAt: "2026-11-22", CounterpartyAccountID: "groceries"},
			{ID: "e2", TransactionID: "txn-1", AccountID: "groceries", Amount: amount, Description: "Groceries", TransactionOccurredAt: "2026-11-22", CounterpartyAccountID: "cash"},
		},
		BalanceDeltas: []entity.BalanceDelta{
			{AccountID: "cash", Amount: testDecimal(t, "-100")},
			{AccountID: "groceries", Amount: amount},
		},
		Idempotency: port.IdempotencyWrite{
			KeyHash: "hash", RequestHash: "fingerprint", TransactionID: "txn-1", CreatedAt: now,
		},
	}

	if err := repo.RecordTransaction(context.Background(), "u1", input); err != nil {
		t.Fatalf("record: %v", err)
	}
	if client.transactInput == nil {
		t.Fatal("expected transaction write")
	}
	if len(client.transactInput.TransactItems) != 9 {
		t.Fatalf("expected 9 writes, got %d", len(client.transactInput.TransactItems))
	}

	var balanceUpdates int
	for _, item := range client.transactInput.TransactItems {
		if item.Update == nil {
			continue
		}
		balanceUpdates++
		if !strings.Contains(item.Update.UpdateExpression, "ADD #balance") {
			t.Fatalf("unexpected balance update: %s", item.Update.UpdateExpression)
		}
	}
	if balanceUpdates != 2 {
		t.Fatalf("expected 2 balance updates, got %d", balanceUpdates)
	}
}

func TestFindTransactionDetailUsesStronglyConsistentQuery(t *testing.T) {
	client := &fakeClient{}
	client.queryItems = []ledgerItem{
		{SK: "TRANSACTION#2026-11-22#txn-1", ID: "txn-1", Description: "Groceries", OccurredAt: "2026-11-22", CreatedAt: "now"},
		{SK: "TRANSACTION#2026-11-22#txn-1#POSTING#p1", ID: "p1", TransactionID: "txn-1", AccountID: "cash", Amount: "-100"},
	}
	repo := NewDynamoDBLedgerRepository(client, "itala")

	transaction, postings, err := repo.FindTransactionDetail(context.Background(), "u1", "2026-11-22", "txn-1")
	if err != nil {
		t.Fatalf("find detail: %v", err)
	}
	if transaction.ID != "txn-1" || len(postings) != 1 {
		t.Fatalf("unexpected result: %+v %+v", transaction, postings)
	}
	if client.queryInput == nil || client.queryInput.ConsistentRead == nil || !*client.queryInput.ConsistentRead {
		t.Fatal("expected strongly consistent query")
	}
}
