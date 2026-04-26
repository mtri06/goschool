package httpx

import (
	"errors"
	"fmt"
	"goschool/internal/service"
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

// ApiErrorMap is a map of standard errors to their corresponding ApiError.
type APIErrorMap map[error]APIError

func (e APIError) WithErr(err error) APIError {
	if err == nil {
		return e
	}
	return APIError{
		Type:   e.Type,
		Msg:    fmt.Sprintf("%s: %v", e.Msg, err),
		Status: e.Status,
	}
}

func (e APIError) WithMsg(msg string) APIError {
	return APIError{
		Type:   e.Type,
		Msg:    fmt.Sprintf("%s: %s", e.Msg, msg),
		Status: e.Status,
	}
}

func RenderError(w http.ResponseWriter, r *http.Request, errMap APIErrorMap, err error) {
	var apiErr APIError

	if errors.As(err, &apiErr) {
		render.Status(r, apiErr.Status)
		render.JSON(w, r, apiErr)
		return
	}

	apiErr = ErrUnknownInternal
	for k, v := range errMap {
		if errors.Is(err, k) {
			apiErr = v
			break
		}
	}

	if apiErr == ErrUnknownInternal {
		log.Error().Err(err).Msg("Unhandled internal error")
	} else {
		var svcErr *service.Error
		if errors.As(err, &svcErr) {
			apiErr.Type = svcErr.Type
		}
		apiErr = apiErr.WithErr(err)
	}

	render.Status(r, apiErr.Status)
	render.JSON(w, r, apiErr)
}

var (
	ErrBadRequest = APIError{
		Type:   "bad_request",
		Msg:    "Bad request",
		Status: http.StatusBadRequest,
	}

	ErrValidationFailed = APIError{
		Type:   "validation_failed",
		Msg:    "Validation failed",
		Status: http.StatusBadRequest,
	}

	ErrInvalidBody = APIError{
		Type:   "invalid_body",
		Msg:    "Invalid request body",
		Status: http.StatusBadRequest,
	}

	ErrInvalidQuery = APIError{
		Type:   "invalid_query",
		Msg:    "Invalid query",
		Status: http.StatusBadRequest,
	}

	ErrInvalidParam = APIError{
		Type:   "invalid_param",
		Msg:    "Invalid URL parameter",
		Status: http.StatusBadRequest,
	}

	ErrUnauthorized = APIError{
		Type:   "unauthorized",
		Msg:    "Unauthorized",
		Status: http.StatusUnauthorized,
	}

	ErrForbidden = APIError{
		Type:   "forbidden",
		Msg:    "Forbidden",
		Status: http.StatusForbidden,
	}

	ErrConflict = APIError{
		Type:   "conflict",
		Msg:    "Conflict",
		Status: http.StatusConflict,
	}

	ErrNotFound = APIError{
		Type:   "not_found",
		Msg:    "Resource not found",
		Status: http.StatusNotFound,
	}

	ErrUnknownInternal = APIError{
		Type:   "internal_error",
		Msg:    "An internal error has occurred",
		Status: http.StatusInternalServerError,
	}
)
