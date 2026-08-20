package api

import (
	"encoding/json"
	"net/http"
)

type response struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Status  int    `json:"status"`
}

func WriteError(w http.ResponseWriter, code string, message string, status int) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response{
		Code:    code,
		Message: message,
		Status:  status,
	})
}
