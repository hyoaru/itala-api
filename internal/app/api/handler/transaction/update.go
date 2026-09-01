package api

import (
	"errors"
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	account "github.com/hyoaru/itala-api/internal/features/account"
	category "github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type updateTransactionRequest struct {
	Amount      string    `json:"amount" validate:"required"`
	AccountID   string    `json:"account_id" validate:"required"`
	CategoryID  string    `json:"category_id" validate:"required"`
	Description string    `json:"description" validate:"required"`
	OccurredAt  time.Time `json:"occurred_at" validate:"required"`
}

func (h *TransactionHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var request updateTransactionRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	amount, err := valueobject.NewDecimal(request.Amount)
	if err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := transaction.UpdateTransactionRequest{
		UserID:      user.ID,
		ID:          r.PathValue("id"),
		Amount:      amount,
		AccountID:   request.AccountID,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt,
	}

	if _, err = h.UpdateTransaction.Execute(r.Context(), useCaseRequest); err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "category not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, account.ErrAccountNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "account not found", http.StatusNotFound)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
