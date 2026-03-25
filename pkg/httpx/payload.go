package httpx

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"
)

var (
	ErrInvalidBody  = errors.New("invalid request body")
	ErrInvalidQuery = errors.New("invalid query parameter type")
)

var validate = validator.New(validator.WithRequiredStructEnabled())

func DecodeBody[T any](r *http.Request) (*T, error) {
	var payload T
	errDecode := render.DecodeJSON(r.Body, &payload)
	errValidate := validate.Struct(payload)
	err := errors.Join(errDecode, errValidate)
	if err != nil {
		return &payload, errors.Join(ErrInvalidBody, err)
	}
	return &payload, nil
}

// GetQueryStr returns the value of the query parameter with the given name.
// If the parameter is not present, it returns an empty string.
func GetQueryStr(r *http.Request, param string) string {
	return r.URL.Query().Get(param)
}

// GetQueryInt returns the value of the query parameter with the given name as an integer.
// If the parameter is not present or cannot be converted to an integer, it returns an error.
func GetQueryInt(r *http.Request, param string) (int, error) {
	str := r.URL.Query().Get(param)
	val, err := strconv.Atoi(str)
	if err != nil {
		return 0, errors.Join(ErrInvalidQuery, err)
	}
	return val, nil
}

// GetQueryArrayStr returns the value of the query parameter with the given name as a slice of strings.
func GetQueryArrayStr(r *http.Request, param string) []string {
	val := r.URL.Query().Get(param)
	return strings.Split(val, ",")
}
