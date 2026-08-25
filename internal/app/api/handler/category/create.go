package api

import (
	"errors"
	"net/http"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobject"
)

type createCategoryRequest struct {
	Name string `json:"name"`
	Type string `json:"type" validate:"omitempty,oneof=INCOME EXPENSE"`
}

type createCategoryResponse struct {
	ID string `json:"id"`
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var request createCategoryRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := category.CreateCategoryRequest{
		UserID: user.ID,
		Name:   request.Name,
		Type:   valueobject.TransactionType(request.Type),
	}

	entity, err := h.CreateCategory.Execute(r.Context(), useCaseRequest)
	if err != nil {
		if errors.Is(err, category.ErrCategoryExists) {
			res.WriteError(w, "RESOURCE_CONFLICT", "category already exists", http.StatusConflict)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	res.WriteJSON(w, http.StatusCreated, createCategoryResponse{ID: entity.ID})
}
