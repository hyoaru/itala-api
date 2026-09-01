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
)

type UpdateTransactionRequest struct {
	UserID         string
	ID             string
	Amount         valueobject.Decimal
	AccountID      string
	CategoryID     string
	Description    string
	OccurredAt     time.Time
	IdempotencyKey string
}

type UpdateTransactionResponse struct{}

type UpdateTransaction struct {
	transactionRepository transactionrepository.TransactionRepository
	categoryRepository    category.CategoryRepository
	accountRepository     account.AccountRepository
}

func NewUpdateTransaction(
	transactionRepository transactionrepository.TransactionRepository,
	categoryRepository category.CategoryRepository,
	accountRepository account.AccountRepository,
) *UpdateTransaction {
	return &UpdateTransaction{
		transactionRepository: transactionRepository,
		categoryRepository:    categoryRepository,
		accountRepository:     accountRepository,
	}
}

func (u *UpdateTransaction) Execute(ctx context.Context, request UpdateTransactionRequest) (UpdateTransactionResponse, error) {
	existing, err := u.transactionRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	foundCategory, err := u.categoryRepository.FindOne(ctx, request.UserID, request.CategoryID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	oldAccount, err := u.accountRepository.FindOne(ctx, request.UserID, existing.AccountID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	newAccount, err := u.accountRepository.FindOne(ctx, request.UserID, request.AccountID)
	if err != nil {
		return UpdateTransactionResponse{}, err
	}

	now := time.Now().UTC()

	transaction := entity.Transaction{
		ID:          request.ID,
		Amount:      request.Amount,
		Type:        foundCategory.TransactionType,
		AccountID:   request.AccountID,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt.UTC(),
		UpdatedAt:   now,
	}

	reverseDelta := existing.Amount
	if existing.Type == valueobject.TransactionTypeIncome {
		reverseDelta = reverseDelta.Negate()
	}
	reverseIdempotencyKey := uuid.New().String()
	if err := u.accountRepository.AdjustBalance(ctx, request.UserID, oldAccount.ID, reverseIdempotencyKey, reverseDelta); err != nil {
		return UpdateTransactionResponse{}, err
	}

	if err := u.transactionRepository.Update(ctx, request.UserID, transaction); err != nil {
		compensateDelta := existing.Amount
		if existing.Type == valueobject.TransactionTypeIncome {
			compensateDelta = compensateDelta.Negate()
		}
		compensateIdempotencyKey := uuid.New().String()
		_ = u.accountRepository.AdjustBalance(ctx, request.UserID, oldAccount.ID, compensateIdempotencyKey, compensateDelta)
		return UpdateTransactionResponse{}, err
	}

	forwardDelta := transaction.Amount
	if transaction.Type == valueobject.TransactionTypeExpense {
		forwardDelta = forwardDelta.Negate()
	}
	forwardIdempotencyKey := uuid.New().String()
	if err := u.accountRepository.AdjustBalance(ctx, request.UserID, newAccount.ID, forwardIdempotencyKey, forwardDelta); err != nil {
		return UpdateTransactionResponse{}, err
	}

	return UpdateTransactionResponse{}, nil
}
