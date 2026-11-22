package accountrepository

import (
	"context"
	"strings"
	"testing"
	"time"

	port "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/external/dynamodbclient"
)

type fakeClient struct {
	transactInput *dynamodbclient.TransactWriteItemsInput
	queryInput    *dynamodbclient.QueryInput
}

func (f *fakeClient) PutItem(context.Context, *dynamodbclient.PutItemInput) error {
	return nil
}

func (f *fakeClient) TransactWriteItems(_ context.Context, input *dynamodbclient.TransactWriteItemsInput) error {
	f.transactInput = input
	return nil
}

func (f *fakeClient) GetItem(_ context.Context, _ *dynamodbclient.GetItemInput, output any) error {
	item := output.(*accountItem)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	*item = accountItem{ID: "a1", Name: "Cash", Type: "asset", CreatedAt: now, UpdatedAt: now}
	return nil
}

func (f *fakeClient) Query(_ context.Context, input *dynamodbclient.QueryInput, output any) (dynamodbclient.QueryOutput, error) {
	f.queryInput = input
	items := output.(*[]accountItem)
	now := time.Now().UTC().Format(time.RFC3339Nano)
	*items = []accountItem{{ID: "a1", Name: "Cash", Type: "asset", CreatedAt: now, UpdatedAt: now}}
	return dynamodbclient.QueryOutput{}, nil
}

func (f *fakeClient) UpdateItem(context.Context, *dynamodbclient.UpdateItemInput) error {
	return nil
}

func (f *fakeClient) DeleteItem(context.Context, *dynamodbclient.DeleteItemInput) error {
	return nil
}

func lockPut(t *testing.T, input *dynamodbclient.TransactWriteItemsInput) map[string]any {
	t.Helper()
	for _, item := range input.TransactItems {
		if item.Put == nil {
			continue
		}
		if sk, ok := item.Put.Item["SK"].(string); ok && strings.HasPrefix(sk, "ACCOUNT_NAME#") {
			return item.Put.Item
		}
	}
	t.Fatal("expected account name lock put")
	return nil
}

func TestCreateWritesAccountAndNameLock(t *testing.T) {
	client := &fakeClient{}
	repo := NewDynamoDBAccountRepository(client, "itala")
	now := time.Now().UTC()

	err := repo.Create(context.Background(), "u1", entity.Account{
		ID:        "a1",
		Name:      " Cash ",
		Type:      entity.TypeAsset,
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if client.transactInput == nil || len(client.transactInput.TransactItems) != 2 {
		t.Fatalf("expected two transactional writes, got %+v", client.transactInput)
	}

	accountPut := client.transactInput.TransactItems[0].Put
	if accountPut == nil || accountPut.Item["SK"] != "ACCOUNT#a1" {
		t.Fatalf("unexpected account put: %+v", accountPut)
	}
	if _, ok := accountPut.Item["balance"]; ok {
		t.Fatal("account item must not carry balance")
	}

	lock := lockPut(t, client.transactInput)
	if lock["SK"] != "ACCOUNT_NAME#cash" || lock["account_id"] != "a1" {
		t.Fatalf("unexpected lock item: %+v", lock)
	}
}

func TestUpdateRenamesNameLock(t *testing.T) {
	client := &fakeClient{}
	repo := NewDynamoDBAccountRepository(client, "itala")

	err := repo.Update(context.Background(), "u1", entity.Account{ID: "a1", Name: "Wallet", Type: entity.TypeAsset})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(client.transactInput.TransactItems) != 3 {
		t.Fatalf("expected account update plus two lock writes, got %d", len(client.transactInput.TransactItems))
	}

	deleted := client.transactInput.TransactItems[1].Delete
	if deleted == nil || deleted.Key["SK"] != "ACCOUNT_NAME#cash" {
		t.Fatalf("unexpected old lock delete: %+v", deleted)
	}
	if deleted.ConditionExpression == nil || *deleted.ConditionExpression != "attribute_exists(PK)" {
		t.Fatal("old lock delete must require the lock to exist")
	}

	lock := lockPut(t, client.transactInput)
	if lock["SK"] != "ACCOUNT_NAME#wallet" {
		t.Fatalf("unexpected new lock: %+v", lock)
	}
}

func TestUpdateSameNormalizedNameSkipsLockWrites(t *testing.T) {
	client := &fakeClient{}
	repo := NewDynamoDBAccountRepository(client, "itala")

	err := repo.Update(context.Background(), "u1", entity.Account{ID: "a1", Name: "  CASH ", Type: entity.TypeAsset})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if len(client.transactInput.TransactItems) != 1 {
		t.Fatalf("expected only account update, got %d", len(client.transactInput.TransactItems))
	}
}

func TestDeleteSoftDeletesAccountAndRemovesLock(t *testing.T) {
	client := &fakeClient{}
	repo := NewDynamoDBAccountRepository(client, "itala")

	err := repo.Delete(context.Background(), "u1", "a1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(client.transactInput.TransactItems) != 2 {
		t.Fatalf("expected account update plus lock delete, got %d", len(client.transactInput.TransactItems))
	}
	if update := client.transactInput.TransactItems[0].Update; update == nil || !strings.Contains(update.UpdateExpression, "deleted_at") {
		t.Fatalf("expected soft delete update, got %+v", update)
	}
	deleted := client.transactInput.TransactItems[1].Delete
	if deleted == nil || deleted.Key["SK"] != "ACCOUNT_NAME#cash" {
		t.Fatalf("unexpected lock delete: %+v", deleted)
	}
	if deleted.ConditionExpression == nil || *deleted.ConditionExpression != "attribute_exists(PK)" {
		t.Fatal("lock delete must require the lock to exist")
	}
}

func TestFindUsesAccountPrefix(t *testing.T) {
	client := &fakeClient{}
	repo := NewDynamoDBAccountRepository(client, "itala")

	page, err := repo.Find(context.Background(), "u1", port.AccountQuery{Limit: 40})
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(page.Accounts) != 1 || page.Accounts[0].Type != entity.TypeAsset {
		t.Fatalf("unexpected page: %+v", page)
	}
	if client.queryInput == nil || client.queryInput.KeyConditionExpression == nil {
		t.Fatal("expected key condition")
	}
	if !strings.Contains(*client.queryInput.KeyConditionExpression, "begins_with(SK") {
		t.Fatalf("unexpected key condition: %s", *client.queryInput.KeyConditionExpression)
	}
	if client.queryInput.FilterExpression == nil || !strings.Contains(*client.queryInput.FilterExpression, "attribute_not_exists(deleted_at)") {
		t.Fatal("expected deleted account filter")
	}
}
