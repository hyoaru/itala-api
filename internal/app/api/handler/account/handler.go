package api

import (
	"github.com/hyoaru/itala-api/internal/features/account"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
)

type AccountHandler struct {
	CreateAccount usecases.UseCase[account.CreateAccountRequest, struct{}]
}
