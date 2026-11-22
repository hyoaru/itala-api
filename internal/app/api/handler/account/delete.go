package api

import (
	"errors"
	"net/http"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

func (h *AccountHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	_, err := h.DeleteAccount.Execute(r.Context(), account.DeleteAccountRequest{
		UserID: user.ID,
		ID:     r.PathValue("id"),
	})
	if err != nil {
		switch {
		case errors.Is(err, account.ErrAccountNotFound):
			res.WriteError(w, "RESOURCE_NOT_FOUND", "account not found", http.StatusNotFound)
		case errors.Is(err, account.ErrConcurrentModification):
			res.WriteError(w, "RESOURCE_CONFLICT", "account changed concurrently", http.StatusConflict)
		default:
			res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
