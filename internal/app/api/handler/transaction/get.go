package api

import (
	"errors"
	"net/http"
	"time"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
)

type getTransactionResponse struct {
	ID          string    `json:"id"`
	Amount      string    `json:"amount"`
	Type        string    `json:"type"`
	AccountID   string    `json:"account_id"`
	CategoryID  string    `json:"category_id"`
	Description string    `json:"description"`
	OccurredAt  time.Time `json:"occurred_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())
	id := r.PathValue("id")

	useCaseRequest := transaction.GetTransactionRequest{
		UserID: user.ID,
		ID:     id,
	}

	entity, err := h.GetTransaction.Execute(r.Context(), useCaseRequest)
	if err != nil {
		if errors.Is(err, transaction.ErrTransactionNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "transaction not found", http.StatusNotFound)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	response := getTransactionResponse{
		ID:          entity.ID,
		Amount:      entity.Amount.String(),
		Type:        string(entity.Type),
		AccountID:   entity.AccountID,
		CategoryID:  entity.CategoryID,
		Description: entity.Description,
		OccurredAt:  entity.OccurredAt,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}

	res.WriteJSON(w, http.StatusOK, response)
}
