package handler

import (
	"goschool/internal/services"
	"goschool/pkg/httpx"
)

func NewErrorMap() httpx.APIErrorMap {
	return map[error]httpx.APIError{
		httpx.ErrInvalidBody:           httpx.ErrBadRequest,
		httpx.ErrInvalidQuery:          httpx.ErrBadRequest,
		services.ErrValidationFailed:   httpx.ErrBadRequest,
		services.ErrNotFound:           httpx.ErrNotFound,
		services.ErrInvalidCredentials: httpx.ErrUnauthorized,
		services.ErrUnauthorized:       httpx.ErrUnauthorized,
	}
}
