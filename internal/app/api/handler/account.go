package api

import (
	"encoding/json"
	"errors"
	"net/http"

	response "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
)

type AccountHandler struct {
	CreateAccount usecases.UseCase[account.CreateAccountRequest, struct{}]
}

type createAccountPayload struct {
	Name string `json:"name"`
}

func (h *AccountHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var payload createAccountPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	request := account.CreateAccountRequest{
		UserID: user.ID,
		Name:   payload.Name,
	}

	if _, err := h.CreateAccount.Execute(r.Context(), request); err != nil {
		if errors.Is(err, account.ErrAccountExists) {
			response.WriteError(w, "RESOURCE_CONFLICT", "account already exists", http.StatusConflict)
			return
		}

		response.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
