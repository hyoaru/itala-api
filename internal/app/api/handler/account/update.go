package api

import (
	"errors"
	"net/http"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

type updateAccountRequest struct {
	Name string `json:"name" validate:"required"`
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var request updateAccountRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	_, err := h.UpdateAccount.Execute(r.Context(), account.UpdateAccountRequest{
		UserID: user.ID,
		ID:     r.PathValue("id"),
		Name:   request.Name,
	})
	if err != nil {
		switch {
		case errors.Is(err, account.ErrInvalidAccount):
			res.WriteError(w, "INVALID_REQUEST", "invalid account", http.StatusBadRequest)
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
