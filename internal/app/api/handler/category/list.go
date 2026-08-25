package api

import (
	"net/http"
	"time"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type listCategoriesRequest struct {
	Limit           int32   `schema:"limit" validate:"omitempty,min=1,max=40"`
	Name            *string `schema:"name"`
	TransactionType *string `schema:"transaction_type" validate:"omitempty,oneof=INCOME EXPENSE"`
	Cursor          *string `schema:"cursor"`
}

type listCategoriesResponseItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	TransactionType string    `json:"transaction_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type listCategoriesResponse struct {
	Items      []listCategoriesResponseItem `json:"items"`
	NextCursor *string                      `json:"next_cursor"`
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	var request listCategoriesRequest
	if err := req.DecodeQuery(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request query", http.StatusBadRequest)
		return
	}

	user := identity.UserFromContext(r.Context())

	limit := request.Limit
	if limit == 0 {
		limit = 40
	}

	var transactionType *valueobject.TransactionType
	if request.TransactionType != nil {
		t := valueobject.TransactionType(*request.TransactionType)
		transactionType = &t
	}

	useCaseRequest := category.ListCategoriesRequest{
		UserID:          user.ID,
		Limit:           limit,
		Name:            request.Name,
		TransactionType: transactionType,
		Cursor:          request.Cursor,
	}

	useCaseResponse, err := h.ListCategories.Execute(r.Context(), useCaseRequest)
	if err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	responseItems := make([]listCategoriesResponseItem, 0, len(useCaseResponse.Categories))
	for _, category := range useCaseResponse.Categories {
		responseItems = append(responseItems, listCategoriesResponseItem{
			ID:              category.ID,
			Name:            category.Name,
			TransactionType: string(category.TransactionType),
			CreatedAt:       category.CreatedAt,
			UpdatedAt:       category.UpdatedAt,
		})
	}

	response := listCategoriesResponse{
		Items:      responseItems,
		NextCursor: useCaseResponse.NextCursor,
	}

	res.WriteJSON(w, http.StatusOK, response)
}
