package transaction

import (
	"context"
	"time"

	"github.com/google/uuid"
	account "github.com/hyoaru/itala-api/internal/features/account"
	category "github.com/hyoaru/itala-api/internal/features/category"
	transactionrepository "github.com/hyoaru/itala-api/internal/features/transaction/application/port/transactionrepository"
	entity "github.com/hyoaru/itala-api/internal/features/transaction/domain/entity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type CreateTransactionRequest struct {
	UserID         string
	Amount         valueobject.Decimal
	AccountID      string
	CategoryID     string
	Description    string
	OccurredAt     time.Time
	IdempotencyKey string
}

type CreateTransactionResponse entity.Transaction

type CreateTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
	categoryRepository    category.CategoryRepository
	accountRepository     account.AccountRepository
	idempotencyStore      idempotency.IdempotencyStore
}

func NewCreateTransaction(
	transactionRepository transactionrepository.TransactionRepository,
	categoryRepository category.CategoryRepository,
	accountRepository account.AccountRepository,
	idempotencyStore idempotency.IdempotencyStore,
) *CreateTransaction {
	return &CreateTransaction{
		transactionRepository: transactionRepository,
		categoryRepository:    categoryRepository,
		accountRepository:     accountRepository,
		idempotencyStore:      idempotencyStore,
	}
}

func (u *CreateTransaction) Execute(ctx context.Context, request CreateTransactionRequest) (CreateTransactionResponse, error) {
	lock, status, _, err := u.idempotencyStore.Acquire(ctx, request.IdempotencyKey, 900)
	if err != nil {
		return CreateTransactionResponse{}, err
	}

	if status == idempotency.IdempotencyStatusLocked {
		return CreateTransactionResponse{}, idempotency.ErrResourceLocked
	}

	if status == idempotency.IdempotencyStatusCompleted {
		return CreateTransactionResponse{}, nil
	}

	defer func() {
		_ = u.idempotencyStore.Commit(ctx, lock, "null")
	}()

	id := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()

	foundCategory, err := u.categoryRepository.FindOne(ctx, request.UserID, request.CategoryID)
	if err != nil {
		return CreateTransactionResponse{}, err
	}

	_, err = u.accountRepository.FindOne(ctx, request.UserID, request.AccountID)
	if err != nil {
		return CreateTransactionResponse{}, err
	}

	transaction := entity.Transaction{
		ID:          id.String(),
		Amount:      request.Amount,
		Type:        foundCategory.TransactionType,
		AccountID:   request.AccountID,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.transactionRepository.Create(ctx, request.UserID, transaction, request.IdempotencyKey); err != nil {
		return CreateTransactionResponse{}, err
	}

	delta := transaction.Amount
	if transaction.Type == valueobject.TransactionTypeExpense {
		delta = delta.Negate()
	}

	balanceIdempotencyKey := uuid.New().String()
	if err := u.accountRepository.AdjustBalance(ctx, request.UserID, request.AccountID, balanceIdempotencyKey, delta); err != nil {
		return CreateTransactionResponse{}, err
	}

	return CreateTransactionResponse(transaction), nil
}
