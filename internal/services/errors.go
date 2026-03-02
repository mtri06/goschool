package services

import (
	"errors"
)

var (
	ErrValidationFailed   = errors.New("validation failed")
	ErrNotFound           = errors.New("resource not found")
	ErrInvalidCredentials = errors.New("invalid username or password")
)
