package account

import (
	"context"
	"fmt"
	"strings"
	"time"

	accountrepository "github.com/hyoaru/itala-api/internal/features/account/application/port/accountrepository"
	entity "github.com/hyoaru/itala-api/internal/features/account/domain/entity"
)

type UpdateAccountRequest struct {
	UserID string
	ID     string
	Name   string
}

type UpdateAccountResponse entity.Account

type UpdateAccount struct {
	accountRepository accountrepository.AccountRepository
}

func NewUpdateAccount(accountRepository accountrepository.AccountRepository) *UpdateAccount {
	return &UpdateAccount{accountRepository: accountRepository}
}

func (u *UpdateAccount) Execute(ctx context.Context, request UpdateAccountRequest) (UpdateAccountResponse, error) {
	if strings.TrimSpace(request.Name) == "" {
		return UpdateAccountResponse{}, fmt.Errorf("%w: name is required", ErrInvalidAccount)
	}

	current, err := u.accountRepository.FindOne(ctx, request.UserID, request.ID)
	if err != nil {
		return UpdateAccountResponse{}, err
	}

	current.Name = request.Name
	current.UpdatedAt = time.Now().UTC()

	if err := u.accountRepository.Update(ctx, request.UserID, current); err != nil {
		return UpdateAccountResponse{}, err
	}

	return UpdateAccountResponse(current), nil
}
