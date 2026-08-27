package api

import (
	"errors"
	"net/http"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

type createAccountRequest struct {
	Name string `json:"name" validate:"required"`
}

type createAccountResponse struct {
	ID string `json:"id"`
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var request createAccountRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := account.CreateAccountRequest{
		UserID: user.ID,
		Name:   request.Name,
	}

	entity, err := h.CreateAccount.Execute(r.Context(), useCaseRequest)
	if err != nil {
		if errors.Is(err, account.ErrAccountExists) {
			res.WriteError(w, "RESOURCE_CONFLICT", "account already exists", http.StatusConflict)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteJSON(w, http.StatusCreated, createAccountResponse{ID: entity.ID})
}
