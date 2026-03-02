package httpx

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type APIError struct {
	// Title is a short, machine-readable string that describes the error.
	Title string `json:"title"`
	// Msg is a human-readable string that describes the error.
	Msg string `json:"message"`
	// Status is the HTTP status code that should be returned with the error.
	Status int `json:"status"`
}

func (e APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Title, e.Msg)
}

// ApiErrorMap is a map of standard errors to their corresponding ApiError.
type APIErrorMap map[error]APIError

func (e APIError) WithErr(err error) APIError {
	if err == nil {
		return e
	}
	return APIError{
		Title:  e.Title,
		Msg:    fmt.Sprintf("%s: %v", e.Msg, err),
		Status: e.Status,
	}
}

func (e APIError) WithMsg(msg string) APIError {
	return APIError{
		Title:  e.Title,
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
		apiErr = apiErr.WithErr(err)
	}

	render.Status(r, apiErr.Status)
	render.JSON(w, r, apiErr)
}

var (
	ErrBadRequest = APIError{
		Title:  "bad_request",
		Msg:    "The request could not be understood or was missing required parameters",
		Status: http.StatusBadRequest,
	}

	ErrUnauthorized = APIError{
		Title:  "unauthorized",
		Msg:    "Authentication failed",
		Status: http.StatusUnauthorized,
	}

	ErrForbidden = APIError{
		Title:  "forbidden",
		Msg:    "You do not have permission to access the requested resource",
		Status: http.StatusForbidden,
	}

	ErrConflict = APIError{
		Title:  "conflict",
		Msg:    "The request could not be completed due to a conflict with the current state of the resource",
		Status: http.StatusConflict,
	}

	ErrNotFound = APIError{
		Title:  "not_found",
		Msg:    "The requested resource could not be found",
		Status: http.StatusNotFound,
	}

	ErrUnknownInternal = APIError{
		Title:  "unknown_internal_error",
		Msg:    "An unknown internal error occurred",
		Status: http.StatusInternalServerError,
	}
)
