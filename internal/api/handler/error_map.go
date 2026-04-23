package handler

import (
	"goschool/internal/service"
	"goschool/pkg/httpx"
)

func NewErrorMap() httpx.APIErrorMap {
	return map[error]httpx.APIError{
		service.ErrValidationFailed: httpx.ErrBadRequest,
		service.ErrNotFound:         httpx.ErrNotFound,
		service.ErrUnauthorized:     httpx.ErrUnauthorized,
		service.ErrForbidden:        httpx.ErrForbidden,
	}
}
