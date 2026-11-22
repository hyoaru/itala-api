package api

import (
	"errors"
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	account "github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/ledger"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type createTransactionRequest struct {
	Amount                string    `json:"amount" validate:"required"`
	AccountID             string    `json:"account_id" validate:"required"`
	CounterpartyAccountID string    `json:"counterparty_account_id" validate:"required"`
	Description           string    `json:"description" validate:"required"`
	OccurredAt            time.Time `json:"occurred_at" validate:"required"`
}

type createTransactionResponse struct {
	ID string `json:"id"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		res.WriteError(w, "INVALID_REQUEST", "Idempotency-Key header is required", http.StatusBadRequest)
		return
	}

	var request createTransactionRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	amount, err := valueobject.NewDecimal(request.Amount)
	if err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid amount", http.StatusBadRequest)
		return
	}

	entity, err := h.CreateTransaction.Execute(r.Context(), ledger.CreateTransactionRequest{
		UserID:                user.ID,
		Amount:                amount,
		AccountID:             request.AccountID,
		CounterpartyAccountID: request.CounterpartyAccountID,
		Description:           request.Description,
		OccurredAt:            request.OccurredAt.UTC(),
		IdempotencyKey:        idempotencyKey,
	})
	if err != nil {
		switch {
		case errors.Is(err, ledger.ErrInvalidTransaction), errors.Is(err, ledger.ErrInvalidCategoryAccount):
			res.WriteError(w, "INVALID_REQUEST", err.Error(), http.StatusBadRequest)
		case errors.Is(err, account.ErrAccountNotFound):
			res.WriteError(w, "RESOURCE_NOT_FOUND", "account not found", http.StatusNotFound)
		case errors.Is(err, ledger.ErrIdempotencyMismatch), errors.Is(err, ledger.ErrConditionFailed):
			res.WriteError(w, "RESOURCE_CONFLICT", "transaction conflict", http.StatusConflict)
		default:
			res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		}
		return
	}

	res.WriteJSON(w, http.StatusCreated, createTransactionResponse{ID: entity.ID})
}
