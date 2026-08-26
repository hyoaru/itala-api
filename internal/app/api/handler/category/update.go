package api

import (
	"net/http"

	req "github.com/hyoaru/itala-api/internal/app/api/request"
	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/category"
	"github.com/hyoaru/itala-api/internal/features/identity"
)

type updateCategoryRequest struct {
	Name   string `json:"name"`
	Status string `json:"status" validate:"oneof=ACTIVE ARCHIVED"`
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())
	id := r.PathValue("id")

	var request updateCategoryRequest
	if err := req.DecodeJSON(r, &request); err != nil {
		res.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	useCaseRequest := category.UpdateCategoryRequest{
		UserID: user.ID,
		ID:     id,
		Name:   request.Name,
		Status: category.Status(request.Status),
	}

	if _, err := h.UpdateCategory.Execute(r.Context(), useCaseRequest); err != nil {
		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
