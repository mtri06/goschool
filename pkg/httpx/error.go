package httpx

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type APIError struct {
	// Type is an machine-readable string that can be used to further categorize the error.
	Type string `json:"type"`
	// Msg is a human-readable string that describes the error.
	Msg string `json:"message"`
	// Status is the HTTP status code that should be returned with the error.
	Status int `json:"status"`
}

func (e APIError) Error() string {
	return e.Msg
}

type appError interface {
	Error() string
	Type() string
	Msg() string
}

// ApiErrorMap is a map of standard errors to their corresponding ApiError.
type APIErrorMap map[error]*APIError

func (e *APIError) WithMsg(msg string) *APIError {
	return &APIError{
		Type:   e.Type,
		Msg:    fmt.Sprintf("%s: %s", e.Msg, msg),
		Status: e.Status,
	}
}

func (e *APIError) Copy() *APIError {
	return &APIError{
		Type:   e.Type,
		Msg:    e.Msg,
		Status: e.Status,
	}
}

// RenderError writes a JSON error response. Resolution order:
//  1. err is (or wraps) an *APIError → render directly.
//  2. err matches a key in errMap via errors.Is → copy the mapped APIError; if err also
//     implements appError, override Msg and Type with the concrete error's values.
//  3. No match → 500 and log.
func RenderError(w http.ResponseWriter, r *http.Request, errMap APIErrorMap, err error) {
	var apiErr *APIError

	// Fast path: err is already a fully-formed *APIError — render it as-is.
	if errors.As(err, &apiErr) {
		render.Status(r, apiErr.Status)
		render.JSON(w, r, apiErr)
		return
	}

	// Walk errMap to find the first sentinel that matches via errors.Is.
	// Copy the mapped value so we can mutate it without affecting the original.
	for k, v := range errMap {
		if !errors.Is(err, k) {
			continue
		}
		apiErr = v.Copy()
		// If the concrete error exposes Msg/Type via the appError interface,
		// use those to give the client a specific message and error type code
		// (e.g. "teacher_not_found") instead of the generic map-level values.
		if aErr, ok := err.(appError); ok {
			// We only shows error messages to client when it's an application error, it should not be an error from library error.
			apiErr = apiErr.WithMsg(aErr.Msg())
			if aErr.Type() != "" {
				apiErr.Type = aErr.Type()
			}
		}
		break
	}

	if apiErr == nil {
		// No match found — log the raw error so it can be investigated.
		log.Error().Err(err).Msg("Unhandled internal error")
		apiErr = ErrUnknownInternal.Copy()
	}

	render.Status(r, apiErr.Status)
	render.JSON(w, r, apiErr)
}

var (
	ErrBadRequest = &APIError{
		Type:   "bad_request",
		Msg:    "Bad request",
		Status: http.StatusBadRequest,
	}

	ErrValidationFailed = &APIError{
		Type:   "validation_failed",
		Msg:    "Validation failed",
		Status: http.StatusBadRequest,
	}

	ErrInvalidBody = &APIError{
		Type:   "invalid_body",
		Msg:    "Invalid request body",
		Status: http.StatusBadRequest,
	}

	ErrInvalidQuery = &APIError{
		Type:   "invalid_query",
		Msg:    "Invalid query",
		Status: http.StatusBadRequest,
	}

	ErrInvalidParam = &APIError{
		Type:   "invalid_param",
		Msg:    "Invalid URL parameter",
		Status: http.StatusBadRequest,
	}

	ErrUnauthorized = &APIError{
		Type:   "unauthorized",
		Msg:    "Unauthorized",
		Status: http.StatusUnauthorized,
	}

	ErrForbidden = &APIError{
		Type:   "forbidden",
		Msg:    "Forbidden",
		Status: http.StatusForbidden,
	}

	ErrConflict = &APIError{
		Type:   "conflict",
		Msg:    "Conflict",
		Status: http.StatusConflict,
	}

	ErrNotFound = &APIError{
		Type:   "not_found",
		Msg:    "Resource not found",
		Status: http.StatusNotFound,
	}

	ErrUnknownInternal = &APIError{
		Type:   "internal_error",
		Msg:    "An internal error has occurred",
		Status: http.StatusInternalServerError,
	}
)
