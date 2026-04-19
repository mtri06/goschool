package httpx

import (
	"errors"
	"net/http"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func DecodeBody[T any](r *http.Request) (*T, error) {
	var payload T
	errDecode := render.DecodeJSON(r.Body, &payload)
	errValidate := validate.Struct(payload)
	err := errors.Join(errDecode, errValidate)
	if err != nil {
		return nil, err
	}
	return &payload, nil
}
