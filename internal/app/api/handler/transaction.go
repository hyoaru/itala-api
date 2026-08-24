package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-playground/validator"
	"github.com/gorilla/schema"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
	"github.com/hyoaru/itala-api/internal/shared/infrastructure/logger"
)

var (
	queryDecoder = schema.NewDecoder()
	validate     = validator.New()
)

type TransactionHandler struct {
	CreateTransaction usecases.UseCase[transaction.CreateTransactionRequest, struct{}]
	ListTransactions  usecases.UseCase[transaction.ListTransactionsRequest, transaction.ListTransactionsResponse]
}

type createTransactionRequest struct {
	Amount      valueobjects.Decimal         `json:"amount"`
	Type        valueobjects.TransactionType `json:"type"`
	CategoryID  string                       `json:"category_id"`
	Description string                       `json:"description"`
	OccurredAt  time.Time                    `json:"occurred_at"`
}

type listTransactionsRequest struct {
	Limit      int32   `schema:"limit" validate:"omitempty,min=1,max=40"`
	Type       *string `schema:"type"`
	CategoryID *string `schema:"category_id"`
	From       *string `schema:"from"`
	To         *string `schema:"to"`
	Cursor     *string `schema:"cursor"`
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

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	transactionType := valueobjects.TransactionType(req.Type)
	if !transactionType.IsValid() {
		res.WriteError(w, "INVALID_REQUEST", "invalid transaction type", http.StatusBadRequest)
		return
	}

	useCaseRequest := transaction.CreateTransactionRequest{
		UserID:      user.ID,
		Amount:      req.Amount,
		Type:        transactionType,
		CategoryID:  req.CategoryID,
		Description: req.Description,
		OccurredAt:  req.OccurredAt,
	}

	if _, err := h.CreateTransaction.Execute(r.Context(), useCaseRequest); err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	var request listTransactionsRequest

	if err := queryDecoder.Decode(&request, r.URL.Query()); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request query", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request query", http.StatusBadRequest)
		return
	}

	logger.Debug("request", request)

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

	useCaseRequest := transaction.ListTransactionsRequest{
		UserID:     user.ID,
		Limit:      limit,
		Type:       transactionType,
		CategoryID: request.CategoryID,
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
