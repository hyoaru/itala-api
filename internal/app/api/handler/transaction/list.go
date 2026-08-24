package api

import (
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type listTransactionsRequest struct {
	Limit      int32      `schema:"limit" validate:"omitempty,min=1,max=40"`
	Type       *string    `schema:"type" validate:"omitempty,oneof=INCOME EXPENSE"`
	CategoryID *string    `schema:"category_id"`
	From       *time.Time `schema:"from"`
	To         *time.Time `schema:"to"`
	Cursor     *string    `schema:"cursor"`
}

type listTransactionsResponseItem struct {
	ID              string    `json:"id"`
	Amount          string    `json:"amount"`
	TransactionType string    `json:"transaction_type"`
	CategoryID      string    `json:"category_id"`
	Description     string    `json:"description"`
	OccurredAt      time.Time `json:"occurred_at"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type listTransactionsResponse struct {
	Items      []listTransactionsResponseItem `json:"items"`
	NextCursor *string                        `json:"next_cursor"`
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	var request listTransactionsRequest
	if err := req.DecodeQuery(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request query", http.StatusBadRequest)
		return
	}

	user := identity.UserFromContext(r.Context())

	limit := request.Limit
	if limit == 0 {
		limit = 40
	}

	var transactionType *valueobjects.TransactionType
	if request.Type != nil {
		t := valueobjects.TransactionType(*request.Type)
		transactionType = &t
	}

	var from *time.Time
	if request.From != nil {
		t := request.From.UTC()
		from = &t
	}

	var to *time.Time
	if request.To != nil {
		t := request.To.UTC()
		to = &t
	}

	useCaseRequest := transaction.ListTransactionsRequest{
		UserID:     user.ID,
		Limit:      limit,
		Type:       transactionType,
		CategoryID: request.CategoryID,
		From:       from,
		To:         to,
		Cursor:     request.Cursor,
	}

	useCaseResponse, err := h.ListTransactions.Execute(r.Context(), useCaseRequest)
	if err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	responseItems := make([]listTransactionsResponseItem, 0, len(useCaseResponse.Transactions))
	for _, transaction := range useCaseResponse.Transactions {
		responseItems = append(responseItems, listTransactionsResponseItem{
			ID:              transaction.ID,
			Amount:          transaction.Amount.String(),
			TransactionType: string(transaction.Type),
			CategoryID:      transaction.CategoryID,
			Description:     transaction.Description,
			OccurredAt:      transaction.OccurredAt,
			CreatedAt:       transaction.CreatedAt,
			UpdatedAt:       transaction.UpdatedAt,
		})
	}

	response := listTransactionsResponse{
		Items:      responseItems,
		NextCursor: useCaseResponse.NextCursor,
	}

	res.WriteJSON(w, http.StatusOK, response)
}
