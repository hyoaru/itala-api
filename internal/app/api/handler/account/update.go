package api

import (
	"net/http"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	"github.com/hyoaru/itala-api/internal/features/identity"
)

type updateAccountRequest struct {
	Name   string `json:"name"`
	Status string `json:"status" validate:"oneof=ACTIVE ARCHIVED"`
}

func (h *AccountHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())
	id := r.PathValue("id")

	var request updateAccountRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := account.UpdateAccountRequest{
		UserID: user.ID,
		ID:     id,
		Name:   request.Name,
		Status: account.Status(request.Status),
	}

	if _, err := h.UpdateAccount.Execute(r.Context(), useCaseRequest); err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
