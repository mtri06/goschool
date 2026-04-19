package handler

import (
	"goschool/internal/services"
	"goschool/pkg/httpx"
)

func NewErrorMap() httpx.APIErrorMap {
	return map[error]httpx.APIError{
		services.ErrValidationFailed: httpx.ErrBadRequest,
		services.ErrNotFound:         httpx.ErrNotFound,
		services.ErrUnauthorized:     httpx.ErrUnauthorized,
	}
}
