package api

import (
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type createTransactionRequest struct {
	Amount      string    `json:"amount"`
	AccountID   string    `json:"account_id"`
	CategoryID  string    `json:"category_id"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
}

type createTransactionResponse struct {
	ID string `json:"id"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var request createTransactionRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	amount, err := valueobjects.NewDecimal(request.Amount)
	if err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := transaction.CreateTransactionRequest{
		UserID:      user.ID,
		Amount:      amount,
		AccountID:   request.AccountID,
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt.UTC(),
	}

	entity, err := h.CreateTransaction.Execute(r.Context(), useCaseRequest)
	if err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteJSON(w, http.StatusCreated, createTransactionResponse{ID: entity.ID})
}
