package account

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

var ErrInvalidAccount = errors.New("invalid account")

type CreateAccountRequest struct {
	UserID string
	Name   string
	Type   entity.Type
}

type CreateAccountResponse entity.Account

type CreateAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewCreateAccount(accountRepository accountrepository.AccountRepository) *CreateAccount {
	return &CreateAccount{accountRepository: accountRepository}
}

func (u *CreateAccount) Execute(ctx context.Context, request CreateAccountRequest) (CreateAccountResponse, error) {
	if strings.TrimSpace(request.Name) == "" || !request.Type.IsValid() {
		return CreateAccountResponse{}, fmt.Errorf("%w: name and type are required", ErrInvalidAccount)
	}

	now := time.Now().UTC()
	account := entity.Account{
		ID:        uuid.Must(uuid.NewV7()).String(),
		Name:      request.Name,
		Type:      request.Type,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := u.accountRepository.Create(ctx, request.UserID, account); err != nil {
		return CreateAccountResponse{}, err
	}

	return CreateAccountResponse(account), nil
}
