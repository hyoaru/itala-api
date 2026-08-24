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
	Amount      valueobjects.Decimal         `json:"amount"`
	Type        valueobjects.TransactionType `json:"type" validate:"omitempty,oneof=INCOME EXPENSE"`
	CategoryID  string                       `json:"category_id"`
	Description string                       `json:"description"`
	OccurredAt  time.Time                    `json:"occurred_at"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var request createTransactionRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := transaction.CreateTransactionRequest{
		UserID:      user.ID,
		Amount:      request.Amount,
		Type:        valueobjects.TransactionType(request.Type),
		CategoryID:  request.CategoryID,
		Description: request.Description,
		OccurredAt:  request.OccurredAt.UTC(),
	}

	if _, err := h.CreateTransaction.Execute(r.Context(), useCaseRequest); err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
