package ledger

import (
	"context"
	"errors"
	"testing"
	"time"

	account "github.com/hyoaru/itala-api/internal/features/account"
	ledgerrepository "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type fakeAccountRepository struct {
	accounts map[string]account.Account
}

func (f *fakeAccountRepository) Create(context.Context, string, account.Account) error { return nil }
func (f *fakeAccountRepository) Find(context.Context, string, account.AccountQuery) (account.AccountPage, error) {
	return account.AccountPage{}, nil
}
func (f *fakeAccountRepository) FindOne(_ context.Context, _ string, id string) (account.Account, error) {
	item, ok := f.accounts[id]
	if !ok {
		return account.Account{}, account.ErrAccountNotFound
	}
	return item, nil
}
func (f *fakeAccountRepository) Update(context.Context, string, account.Account) error { return nil }
func (f *fakeAccountRepository) Delete(context.Context, string, string) error          { return nil }

type fakeLedgerRepository struct {
	input       ledgerrepository.RecordTransactionInput
	recorded    bool
	idempotency map[string]ledgerrepository.IdempotencyRecord
}

func (f *fakeLedgerRepository) RecordTransaction(_ context.Context, _ string, input ledgerrepository.RecordTransactionInput) error {
	f.input = input
	f.recorded = true
	f.idempotency[input.Idempotency.KeyHash] = ledgerrepository.IdempotencyRecord{
		TransactionID: input.Idempotency.TransactionID,
		RequestHash:   input.Idempotency.RequestHash,
		CreatedAt:     input.Idempotency.CreatedAt,
	}
	return nil
}

func (f *fakeLedgerRepository) FindIdempotency(_ context.Context, _ string, keyHash string) (ledgerrepository.IdempotencyRecord, bool, error) {
	record, ok := f.idempotency[keyHash]
	return record, ok, nil
}

func (f *fakeLedgerRepository) FindTransactionDetail(context.Context, string, string, string) (entity.Transaction, []entity.Posting, error) {
	return entity.Transaction{}, nil, ledgerrepository.ErrTransactionNotFound
}

func newFixtures() (*fakeAccountRepository, *fakeLedgerRepository, CreateTransactionRequest) {
	accounts := &fakeAccountRepository{accounts: map[string]account.Account{
		"cash":      {ID: "cash", Name: "Cash", Type: account.TypeAsset},
		"groceries": {ID: "groceries", Name: "Groceries", Type: account.TypeExpense},
	}}
	ledgerRepo := &fakeLedgerRepository{idempotency: map[string]ledgerrepository.IdempotencyRecord{}}
	amount, _ := valueobject.NewDecimal("100")
	request := CreateTransactionRequest{
		UserID:                "u1",
		Amount:                amount,
		AccountID:             "cash",
		CounterpartyAccountID: "groceries",
		Description:           "Weekly groceries",
		OccurredAt:            time.Date(2026, 11, 22, 10, 30, 0, 0, time.UTC),
		IdempotencyKey:        "key-1",
	}
	return accounts, ledgerRepo, request
}

func TestCreateTransactionWritesBalancedPostings(t *testing.T) {
	accounts, ledgerRepo, request := newFixtures()
	useCase := NewCreateTransaction(ledgerRepo, accounts)

	result, err := useCase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if result.ID == "" || !ledgerRepo.recorded {
		t.Fatal("expected transaction to be recorded")
	}
	if len(ledgerRepo.input.Postings) != 2 || len(ledgerRepo.input.BalanceDeltas) != 2 {
		t.Fatalf("unexpected postings or deltas: %+v", ledgerRepo.input)
	}
	sum, _ := valueobject.NewDecimal("0")
	for _, posting := range ledgerRepo.input.Postings {
		sum = sum.Add(posting.Amount)
	}
	if !sum.IsZero() {
		t.Fatalf("postings must sum to zero, got %s", sum.String())
	}
	if ledgerRepo.input.Summary.DisplayAccountID != "groceries" || ledgerRepo.input.Summary.DisplayAmount.String() != "100" {
		t.Fatalf("unexpected summary: %+v", ledgerRepo.input.Summary)
	}
}

func TestCreateTransactionReplaysIdempotency(t *testing.T) {
	accounts, ledgerRepo, request := newFixtures()
	useCase := NewCreateTransaction(ledgerRepo, accounts)

	first, err := useCase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("first execute: %v", err)
	}
	second, err := useCase.Execute(context.Background(), request)
	if err != nil {
		t.Fatalf("second execute: %v", err)
	}
	if first.ID != second.ID {
		t.Fatalf("replay returned different id: %s vs %s", first.ID, second.ID)
	}

	changed := request
	amount, _ := valueobject.NewDecimal("200")
	changed.Amount = amount
	if _, err := useCase.Execute(context.Background(), changed); !errors.Is(err, ledgerrepository.ErrIdempotencyMismatch) {
		t.Fatalf("expected idempotency mismatch, got %v", err)
	}
}

func TestCreateTransactionRejectsInvalidInput(t *testing.T) {
	accounts, ledgerRepo, request := newFixtures()
	useCase := NewCreateTransaction(ledgerRepo, accounts)

	zero, _ := valueobject.NewDecimal("0")
	request.Amount = zero
	if _, err := useCase.Execute(context.Background(), request); !errors.Is(err, ErrInvalidTransaction) {
		t.Fatalf("expected invalid transaction, got %v", err)
	}

	_, _, freshRequest := newFixtures()
	freshRequest.CounterpartyAccountID = "cash"
	if _, err := useCase.Execute(context.Background(), freshRequest); !errors.Is(err, ErrInvalidCategoryAccount) {
		t.Fatalf("expected invalid category account, got %v", err)
	}
}
