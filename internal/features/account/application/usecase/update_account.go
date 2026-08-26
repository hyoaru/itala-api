package account

import (
	"context"
	"time"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
	valueobject "github.com/hyoaru/itala-api/internal/features/account/domain/valueobject"
)

type UpdateAccountRequest struct {
	UserID string
	ID     string
	Name   string
	Status valueobject.Status
}

type UpdateAccountResponse struct{}

type UpdateAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewUpdateAccount(accountRepository accountrepository.AccountRepository) *UpdateAccount {
	return &UpdateAccount{accountRepository: accountRepository}
}

func (u *UpdateAccount) Execute(ctx context.Context, request UpdateAccountRequest) (UpdateAccountResponse, error) {
	now := time.Now().UTC()

	account := entity.Account{
		ID:        request.ID,
		Name:      request.Name,
		Status:    request.Status,
		UpdatedAt: now,
	}

	if err := u.accountRepository.Update(ctx, request.UserID, account); err != nil {
		return UpdateAccountResponse{}, err
	}

	return UpdateAccountResponse{}, nil
}
