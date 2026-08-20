package api

import (
	"encoding/json"
	"errors"
	"net/http"

	response "github.com/hyoaru/itala-api/internal/app/api/response"
	"github.com/hyoaru/itala-api/internal/features/category"
	identity "github.com/hyoaru/itala-api/internal/features/identity"
	"github.com/hyoaru/itala-api/internal/shared/application/usecases"
	"github.com/hyoaru/itala-api/internal/shared/domain/valueobjects"
)

type CategoryHandler struct {
	CreateCategory usecases.UseCase[category.CreateCategoryRequest, struct{}]
}

type createCategoryPayload struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	user := identity.UserFromContext(r.Context())

	var payload createCategoryPayload
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.WriteError(w, "INVALID_REQUEST", "invalid request body", http.StatusBadRequest)
		return
	}

	transactionType := valueobjects.TransactionType(payload.Type)
	if !transactionType.IsValid() {
		response.WriteError(w, "INVALID_REQUEST", "invalid transaction type", http.StatusBadRequest)
		return
	}

	request := category.CreateCategoryRequest{
		UserID:          user.ID,
		Name:            payload.Name,
		TransactionType: transactionType,
	}

	if _, err := h.CreateCategory.Execute(r.Context(), request); err != nil {
		if errors.Is(err, category.ErrCategoryExists) {
			response.WriteError(w, "RESOURCE_CONFLICT", "category already exists", http.StatusConflict)
			return
		}

		response.WriteError(w, "INTERNAL_SERVER_ERROR", "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}
