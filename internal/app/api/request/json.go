package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator"
)

func DecodeJSON(r *http.Request, output any) error {
	validate := validator.New()
	if err := json.NewDecoder(r.Body).Decode(&output); err != nil {
		return err
	}

	return validate.Struct(output)
}
