package api

import (
	"encoding/json"
	"net/http"
	"time"

	response "github.com/hyoaru/itala-api/internal/app/api/response"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/features/transaction"
	entities "github.com/hyoaru/itala-api/internal/features/transaction/domain/entities"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type TransactionHandler struct {
	CreateTransaction usecases.UseCase[transaction.CreateTransactionRequest, struct{}]
	ListTransactions  usecases.UseCase[transaction.ListTransactionsRequest, []entities.Transaction]
}

type createTransactionRequest struct {
	Amount      valueobjects.Decimal         `json:"amount"`
	Type        valueobjects.TransactionType `json:"type"`
	CategoryID  string                       `json:"category_id"`
	Description string                       `json:"description"`
	OccurredAt  time.Time                    `json:"occurred_at"`
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

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var req createTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	transactionType := valueobjects.TransactionType(req.Type)
	if !transactionType.IsValid() {
		response.WriteError(w, "INVALID_REQUEST", "invalid transaction type", http.StatusBadRequest)
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
		response.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *TransactionHandler) List(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	req := transaction.ListTransactionsRequest{
		UserID: user.ID,
	}

	transactions, err := h.ListTransactions.Execute(r.Context(), req)
	if err != nil {
		response.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	res := make([]listTransactionsResponseItem, 0, len(transactions))
	for _, transaction := range transactions {
		res = append(res, listTransactionsResponseItem{
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

	response.WriteJSON(w, http.StatusOK, res)
}
