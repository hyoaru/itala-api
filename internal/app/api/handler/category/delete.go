package api

import (
	"errors"
	"net/http"

	res "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	useCaseRequest := category.DeleteCategoryRequest{
		UserID: user.ID,
		ID:     r.PathValue("id"),
	}

	if _, err := h.DeleteCategory.Execute(r.Context(), useCaseRequest); err != nil {
		if errors.Is(err, category.ErrCategoryNotFound) {
			res.WriteError(w, "RESOURCE_NOT_FOUND", "category not found", http.StatusNotFound)
			return
		}

		res.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
