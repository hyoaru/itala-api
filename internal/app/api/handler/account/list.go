package api

import (
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

type listAccountsRequest struct {
	Limit  int32   `schema:"limit" validate:"omitempty,min=1,max=40"`
	Name   *string `schema:"name"`
	Cursor *string `schema:"cursor"`
}

type listAccountsResponseItem struct {
	ID        string              `json:"id"`
	Name      string              `json:"name"`
	Type      account.AccountType `json:"type"`
	CreatedAt time.Time           `json:"created_at"`
	UpdatedAt time.Time           `json:"updated_at"`
}

type listAccountsResponse struct {
	Items      []listAccountsResponseItem `json:"items"`
	NextCursor *string                    `json:"next_cursor"`
}

func (h *AccountHandler) List(w http.ResponseWriter, r *http.Request) {
	var request listAccountsRequest
	if err := req.DecodeQuery(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request query", http.StatusBadRequest)
		return
	}

	user := identity.UserFromContext(r.Context())
	limit := request.Limit
	if limit == 0 {
		limit = 40
	}

	page, err := h.ListAccounts.Execute(r.Context(), account.ListAccountsRequest{
		UserID: user.ID,
		Limit:  limit,
		Name:   request.Name,
		Cursor: request.Cursor,
	})
	if err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	items := make([]listAccountsResponseItem, 0, len(page.Accounts))
	for _, item := range page.Accounts {
		items = append(items, listAccountsResponseItem{
			ID:        item.ID,
			Name:      item.Name,
			Type:      item.Type,
			CreatedAt: item.CreatedAt,
			UpdatedAt: item.UpdatedAt,
		})
	}

	res.WriteJSON(w, http.StatusOK, listAccountsResponse{Items: items, NextCursor: page.NextCursor})
}
