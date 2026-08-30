package api

import (
	"errors"
	"net/http"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

func (h *AccountHandler) Restore(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	useCaseRequest := account.RestoreAccountRequest{
		UserID: user.ID,
		ID:     r.PathValue("id"),
	}

	if _, err := h.RestoreAccount.Execute(r.Context(), useCaseRequest); err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "account not found", http.StatusNotFound)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
