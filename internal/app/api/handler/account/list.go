package api

import (
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/account"
	"github.com/hyoaru/itala-api/internal/features/identity"
)

type listAccountsRequest struct {
	Limit  int32   `schema:"limit" validate:"omitempty,min=1,max=40"`
	Name   *string `schema:"name"`
	Cursor *string `schema:"cursor"`
}

type listAccountsResponseItem struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Balance   string    `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
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

	useCaseRequest := account.ListAccountsRequest{
		UserID: user.ID,
		Limit:  limit,
		Name:   request.Name,
		Cursor: request.Cursor,
	}

	useCaseResponse, err := h.ListAccounts.Execute(r.Context(), useCaseRequest)
	if err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	responseItems := make([]listAccountsResponseItem, 0, len(useCaseResponse.Accounts))
	for _, account := range useCaseResponse.Accounts {
		responseItems = append(responseItems, listAccountsResponseItem{
			ID:        account.ID,
			Name:      account.Name,
			Balance:   account.Balance.String(),
			CreatedAt: account.CreatedAt,
			UpdatedAt: account.UpdatedAt,
		})
	}

	response := listAccountsResponse{
		Items:      responseItems,
		NextCursor: useCaseResponse.NextCursor,
	}

	res.WriteJSON(w, http.StatusOK, response)
}
