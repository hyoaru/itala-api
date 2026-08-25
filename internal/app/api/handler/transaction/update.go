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
	Amount      string    `json:"amount"`
	AccountID   string    `json:"account_id"`
	CategoryID  string    `json:"category_id"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
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
		if errors.Is(err, category.ErrCategoryArchived) {
			res.WriteError(w, "RESOURCE_CONFLICT", "category is archived", http.StatusConflict)
			return
		}

		if errors.Is(err, account.ErrAccountArchived) {
			res.WriteError(w, "RESOURCE_CONFLICT", "account is archived", http.StatusConflict)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
