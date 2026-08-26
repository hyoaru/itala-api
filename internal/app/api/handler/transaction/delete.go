package api

import (
	"errors"
	"net/http"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
)

func (h *TransactionHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	useCaseRequest := transaction.DeleteTransactionRequest{
		UserID: user.ID,
		ID:     r.PathValue("id"),
	}

	if _, err := h.DeleteTransaction.Execute(r.Context(), useCaseRequest); err != nil {
		if errors.Is(err, transaction.ErrTransactionNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "transaction not found", http.StatusNotFound)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
