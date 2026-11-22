package api

import (
	"errors"
	"net/http"
	"time"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

type getAccountResponse struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Type      account.AccountType `json:"type"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

func (h *AccountHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	entity, err := h.GetAccount.Execute(r.Context(), account.GetAccountRequest{
		UserID: user.ID,
		ID:     r.PathValue("id"),
	})
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "account not found", http.StatusNotFound)
			return
		}
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteJSON(w, http.StatusOK, getAccountResponse{
		ID:        entity.ID,
		Name:      entity.Name,
		Type:      entity.Type,
		CreatedAt: entity.CreatedAt,
		UpdatedAt: entity.UpdatedAt,
	})
}
