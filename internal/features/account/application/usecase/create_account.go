package account

import (
	"context"
	"time"

	"github.com/google/uuid"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	accountvalueobject "github.com/hyoaru/itala-api/internal/features/account/domain/valueobject"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type CreateAccountRequest struct {
	UserID          string
	Name            string
	TransactionType valueobject.TransactionType
}

type CreateAccountResponse entity.Account

type CreateAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewCreateAccount(accountRepository accountrepository.AccountRepository) *CreateAccount {
	return &CreateAccount{accountRepository: accountRepository}
}

func (u *CreateAccount) Execute(ctx context.Context, request CreateAccountRequest) (CreateAccountResponse, error) {
	id := uuid.Must(uuid.NewV7())
	now := time.Now().UTC()
	balance, err := valueobject.NewDecimal("0")
	if err != nil {
		return CreateAccountResponse{}, err
	}

	account := entity.Account{
		ID:        id.String(),
		Name:      request.Name,
		Balance:   balance,
		Status:    accountvalueobject.StatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.accountRepository.Create(ctx, request.UserID, account); err != nil {
		return CreateAccountResponse{}, err
	}

	return CreateAccountResponse(account), nil
}
