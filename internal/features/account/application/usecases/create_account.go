package account

import (
	"context"
	"time"

	"github.com/google/uuid"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/ports/accountrepository"
	entities "github.com/hyoaru/itala-api/internal/features/account/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/shopspring/decimal"
)

type CreateAccountRequest struct {
	UserID          string
	Name            string
	TransactionType valueobjects.TransactionType
}

type CreateAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewCreateAccount(accountRepository accountrepository.AccountRepository) *CreateAccount {
	return &CreateAccount{accountRepository: accountRepository}
}

func (u *CreateAccount) Execute(ctx context.Context, request CreateAccountRequest) (struct{}, error) {
	id := uuid.New()
	now := time.Now().UTC()

	account := entities.Account{
		ID:        id.String(),
		Name:      request.Name,
		Balance:   decimal.Zero,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.accountRepository.Create(ctx, request.UserID, account); err != nil {
		return struct{}{}, err
	}

	return struct{}{}, nil
}
