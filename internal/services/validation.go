package services

import (
	"fmt"
	"unicode"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

// validatePassword checks that the password meets the following requirements:
// - at least 8 characters long
// - at most 72 characters long
// - contains at least one uppercase letter
// - contains at least one lowercase letter
// - contains at least one digit
// - contains at least one special character
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters long", ErrValidationFailed, minPasswordLength)
	}

	if len(password) > maxPasswordLength {
		return fmt.Errorf("%w: password must be at most %d characters long", ErrValidationFailed, maxPasswordLength)
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return fmt.Errorf("%w: password must contain at least one uppercase letter", ErrValidationFailed)
	}
	if !hasLower {
		return fmt.Errorf("%w: password must contain at least one lowercase letter", ErrValidationFailed)
	}
	if !hasDigit {
		return fmt.Errorf("%w: password must contain at least one digit", ErrValidationFailed)
	}
	if !hasSpecial {
		return fmt.Errorf("%w: password must contain at least one special character", ErrValidationFailed)
	}

	return nil
}
