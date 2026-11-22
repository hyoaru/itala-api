package api

import (
	"errors"
	"net/http"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/ledger"
)

type getTransactionQuery struct {
	OccurredAt string `schema:"occurred_at" validate:"required"`
}

type getTransactionResponse struct {
	ID                      string                  `json:"id"`
	Description             string                  `json:"description"`
	OccurredAt              string                  `json:"occurred_at"`
	CreatedAt               string                  `json:"created_at"`
	ReversesTransactionID   *string                 `json:"reverses_transaction_id,omitempty"`
	ReversedByTransactionID *string                 `json:"reversed_by_transaction_id,omitempty"`
	Postings                []getTransactionPosting `json:"postings"`
}

type getTransactionPosting struct {
	ID        string `json:"id"`
	AccountID string `json:"account_id"`
	Amount    string `json:"amount"`
}

func (h *TransactionHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var query getTransactionQuery
	if err := req.DecodeQuery(r, &query); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "occurred_at is required", http.StatusBadRequest)
		return
	}

	result, err := h.GetTransaction.Execute(r.Context(), ledger.GetTransactionRequest{
		UserID:     user.ID,
		OccurredAt: query.OccurredAt,
		ID:         r.PathValue("id"),
	})
	if err != nil {
		if errors.Is(err, ledger.ErrTransactionNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "transaction not found", http.StatusNotFound)
			return
		}
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	postings := make([]getTransactionPosting, 0, len(result.Postings))
	for _, posting := range result.Postings {
		postings = append(postings, getTransactionPosting{
			ID:        posting.ID,
			AccountID: posting.AccountID,
			Amount:    posting.Amount.String(),
		})
	}

	res.WriteJSON(w, http.StatusOK, getTransactionResponse{
		ID:                      result.Transaction.ID,
		Description:             result.Transaction.Description,
		OccurredAt:              result.Transaction.OccurredAt,
		CreatedAt:               result.Transaction.CreatedAt,
		ReversesTransactionID:   result.Transaction.ReversesTransactionID,
		ReversedByTransactionID: result.Transaction.ReversedByTransactionID,
		Postings:                postings,
	})
}
