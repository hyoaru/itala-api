package api

import (
	"fmt"
	"net/http"

	"github.com/hyoaru/itala-api/internal/shared/logger"
)

type TransactionHandler struct{}

func (h *TransactionHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	authorization := r.Header.Get("Authorization")
	logger.Debug("Received auth", map[string]any{"auth": authorization})
	w.Write([]byte(fmt.Sprintf(`{"auth": "%s"}`, authorization)))
}
