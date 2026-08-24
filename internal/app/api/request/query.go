package api

import (
	"net/http"

	"github.com/go-playground/validator"
	"github.com/gorilla/schema"
)

func DecodeQuery(r *http.Request, output any) error {
	var (
		decoder  = schema.NewDecoder()
		validate = validator.New()
	)

	if err := decoder.Decode(output, r.URL.Query()); err != nil {
		return err
	}

	return validate.Struct(output)
}
