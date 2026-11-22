package ledger

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	account "github.com/hyoaru/itala-api/internal/features/account"
	ledgerrepository "github.com/hyoaru/itala-api/internal/features/ledger/application/port/ledgerrepository"
	entity "github.com/hyoaru/itala-api/internal/features/ledger/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

var (
	ErrInvalidTransaction     = errors.New("invalid transaction")
	ErrInvalidCategoryAccount = errors.New("counterparty account must be income or expense")
)

type CreateTransactionRequest struct {
	UserID                string
	Amount                valueobject.Decimal
	AccountID             string
	CounterpartyAccountID string
	Description           string
	OccurredAt            time.Time
	IdempotencyKey        string
}

type CreateTransactionResponse struct {
	ID string `json:"id"`
}

type CreateTransaction struct {
	ledgerRepository  ledgerrepository.LedgerRepository
	accountRepository account.AccountRepository
}

func NewCreateTransaction(
	ledgerRepository ledgerrepository.LedgerRepository,
	accountRepository account.AccountRepository,
) *CreateTransaction {
	return &CreateTransaction{ledgerRepository: ledgerRepository, accountRepository: accountRepository}
}

type requestFingerprint struct {
	Amount                string `json:"amount"`
	AccountID             string `json:"account_id"`
	CounterpartyAccountID string `json:"counterparty_account_id"`
	Description           string `json:"description"`
	OccurredAt            string `json:"occurred_at"`
}

func keyHash(userID string, key string) string {
	sum := sha256.Sum256([]byte(userID + "|transaction.create|" + key))
	return hex.EncodeToString(sum[:])
}

func requestHash(request CreateTransactionRequest) string {
	fingerprint := requestFingerprint{
		Amount:                request.Amount.String(),
		AccountID:             request.AccountID,
		CounterpartyAccountID: request.CounterpartyAccountID,
		Description:           request.Description,
		OccurredAt:            request.OccurredAt.UTC().Format(time.RFC3339Nano),
	}
	encoded, err := json.Marshal(fingerprint)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}

func (u *CreateTransaction) resolveReplay(ctx context.Context, userID string, hash string, fingerprint string) (string, error) {
	record, found, err := u.ledgerRepository.FindIdempotency(ctx, userID, hash)
	if err != nil {
		return "", err
	}
	if !found {
		return "", nil
	}
	if record.RequestHash != fingerprint {
		return "", ledgerrepository.ErrIdempotencyMismatch
	}
	return record.TransactionID, nil
}

func (u *CreateTransaction) Execute(ctx context.Context, request CreateTransactionRequest) (CreateTransactionResponse, error) {
	if request.Amount.IsZero() || request.Amount.IsNegative() {
		return CreateTransactionResponse{}, fmt.Errorf("%w: amount must be positive", ErrInvalidTransaction)
	}
	if strings.TrimSpace(request.Description) == "" {
		return CreateTransactionResponse{}, fmt.Errorf("%w: description is required", ErrInvalidTransaction)
	}
	if request.AccountID == "" || request.CounterpartyAccountID == "" || request.IdempotencyKey == "" {
		return CreateTransactionResponse{}, fmt.Errorf("%w: account, counterparty, and idempotency key are required", ErrInvalidTransaction)
	}

	balanceAccount, err := u.accountRepository.FindOne(ctx, request.UserID, request.AccountID)
	if err != nil {
		return CreateTransactionResponse{}, err
	}
	counterparty, err := u.accountRepository.FindOne(ctx, request.UserID, request.CounterpartyAccountID)
	if err != nil {
		return CreateTransactionResponse{}, err
	}
	if counterparty.Type != account.TypeIncome && counterparty.Type != account.TypeExpense {
		return CreateTransactionResponse{}, ErrInvalidCategoryAccount
	}

	hash := keyHash(request.UserID, request.IdempotencyKey)
	fingerprint := requestHash(request)
	existing, err := u.resolveReplay(ctx, request.UserID, hash, fingerprint)
	if err != nil {
		return CreateTransactionResponse{}, err
	}
	if existing != "" {
		return CreateTransactionResponse{ID: existing}, nil
	}

	occurredAt := request.OccurredAt.UTC().Format("2006-01-02")
	createdAt := time.Now().UTC().Format(time.RFC3339Nano)
	transactionID := uuid.Must(uuid.NewV7()).String()

	balanceAmount := request.Amount.Negate()
	if counterparty.Type == account.TypeIncome {
		balanceAmount = request.Amount
	}
	counterpartyAmount := balanceAmount.Negate()

	balancePosting := entity.Posting{
		ID:            uuid.Must(uuid.NewV7()).String(),
		TransactionID: transactionID,
		AccountID:     balanceAccount.ID,
		Amount:        balanceAmount,
	}
	counterpartyPosting := entity.Posting{
		ID:            uuid.Must(uuid.NewV7()).String(),
		TransactionID: transactionID,
		AccountID:     counterparty.ID,
		Amount:        counterpartyAmount,
	}
	postings := []entity.Posting{balancePosting, counterpartyPosting}

	statementEntries := make([]entity.AccountStatementEntry, 0, len(postings))
	balanceDeltas := make([]entity.BalanceDelta, 0, len(postings))
	for _, posting := range postings {
		counterpartyID := balanceAccount.ID
		if posting.AccountID == balanceAccount.ID {
			counterpartyID = counterparty.ID
		}
		statementEntries = append(statementEntries, entity.AccountStatementEntry{
			ID:                    uuid.Must(uuid.NewV7()).String(),
			TransactionID:         transactionID,
			AccountID:             posting.AccountID,
			Amount:                posting.Amount,
			Description:           request.Description,
			TransactionOccurredAt: occurredAt,
			CounterpartyAccountID: counterpartyID,
		})
		balanceDeltas = append(balanceDeltas, entity.BalanceDelta{AccountID: posting.AccountID, Amount: posting.Amount})
	}

	input := ledgerrepository.RecordTransactionInput{
		TransactionID: transactionID,
		Description:   request.Description,
		OccurredAt:    occurredAt,
		CreatedAt:     createdAt,
		Postings:      postings,
		Summary: entity.TransactionSummary{
			ID:                    transactionID,
			TransactionID:         transactionID,
			Description:           request.Description,
			OccurredAt:            occurredAt,
			DisplayAmount:         counterpartyAmount,
			DisplayAccountID:      counterparty.ID,
			CounterpartyAccountID: balanceAccount.ID,
		},
		StatementEntries: statementEntries,
		BalanceDeltas:    balanceDeltas,
		Idempotency: ledgerrepository.IdempotencyWrite{
			KeyHash:       hash,
			RequestHash:   fingerprint,
			TransactionID: transactionID,
			CreatedAt:     createdAt,
		},
	}

	if err := u.ledgerRepository.RecordTransaction(ctx, request.UserID, input); err != nil {
		if errors.Is(err, ledgerrepository.ErrConditionFailed) {
			existing, replayErr := u.resolveReplay(ctx, request.UserID, hash, fingerprint)
			if replayErr != nil {
				return CreateTransactionResponse{}, replayErr
			}
			if existing != "" {
				return CreateTransactionResponse{ID: existing}, nil
			}
		}
		return CreateTransactionResponse{}, err
	}

	return CreateTransactionResponse{ID: transactionID}, nil
}
