package api

import (
	"errors"
	"net/http"
	"time"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

type getCategoryResponse struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	TransactionType string    `json:"transaction_type"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())
	id := r.PathValue("id")

	useCaseRequest := category.GetCategoryRequest{
		UserID: user.ID,
		ID:     id,
	}

	entity, err := h.GetCategory.Execute(r.Context(), useCaseRequest)
	if err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "category not found", http.StatusNotFound)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	response := getCategoryResponse{
		ID:              entity.ID,
		Name:            entity.Name,
		TransactionType: string(entity.TransactionType),
		Status:          string(entity.Status),
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
	}

	res.WriteJSON(w, http.StatusOK, response)
}
