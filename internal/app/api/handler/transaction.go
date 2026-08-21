package api

import (
	"encoding/json"
	"net/http"
	"time"

	response "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type TransactionHandler struct {
	CreateTransaction usecases.UseCase[transaction.CreateTransactionRequest, struct{}]
}

type createTransactionPayload struct {
	Amount      valueobjects.Decimal         `json:"amount"`
	Type        valueobjects.TransactionType `json:"type"`
	CategoryID  string                       `json:"category_id"`
	Description string                       `json:"description"`
	OccurredAt  time.Time                    `json:"occurred_at"`
}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var payload createTransactionPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	transactionType := valueobjects.TransactionType(payload.Type)
	if !transactionType.IsValid() {
		response.WriteError(w, "INVALID_REQUEST", "invalid transaction type", http.StatusBadRequest)
		return
	}

	request := transaction.CreateTransactionRequest{
		UserID:      user.ID,
		Amount:      payload.Amount,
		Type:        transactionType,
		CategoryID:  payload.CategoryID,
		Description: payload.Description,
		OccurredAt:  payload.OccurredAt,
	}

	if _, err := h.CreateTransaction.Execute(r.Context(), request); err != nil {
		response.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
