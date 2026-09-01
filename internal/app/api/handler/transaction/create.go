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
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/idempotency"
)

type createTransactionRequest struct {
	Amount      string    `json:"amount" validate:"required"`
	AccountID   string    `json:"account_id" validate:"required"`
	CategoryID  string    `json:"category_id" validate:"required"`
	Description string    `json:"description" validate:"required"`
	OccurredAt  time.Time `json:"occurred_at" validate:"required"`
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
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := transaction.CreateTransactionRequest{
		UserID:         user.ID,
		Amount:         amount,
		AccountID:      request.AccountID,
		CategoryID:     request.CategoryID,
		Description:    request.Description,
		OccurredAt:     request.OccurredAt.UTC(),
		IdempotencyKey: idempotencyKey,
	}

	entity, err := h.CreateTransaction.Execute(r.Context(), useCaseRequest)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "category not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, account.ErrAccountNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "account not found", http.StatusNotFound)
			return
		}

		if errors.Is(err, idempotency.ErrResourceLocked) {
			res.WriteError(w, "RESOURCE_LOCKED", "operation in progress", http.StatusConflict)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteJSON(w, http.StatusCreated, createTransactionResponse{ID: entity.ID})
}
