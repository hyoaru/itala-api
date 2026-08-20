package api

import (
	"fmt"
	"net/http"

	identity "github.com/hyoaru/itala-api/internal/features/identity"
)

type TransactionHandler struct{}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user := identity.UserFromContext(r.Context())
	w.Write([]byte(fmt.Sprintf(`{"user": "%s"}`, user.ID)))
}
