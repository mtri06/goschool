package service

import (
	"fmt"
	"slices"
	"unicode"
)

// ----------------------------------
// Common validation functions used across multiple services. These are not specific to any one service
// ----------------------------------

func validateGender(gender string) error {
	if !slices.Contains(allGenders, gender) {
		return fmt.Errorf("gender must be one of %v", allGenders)
	}
	return nil
}

func validateRoles(role string) error {
	if !slices.Contains(allRoles, role) {
		return fmt.Errorf("role must be one of %v", allRoles)
	}
	return nil
}

func validateTeacherWorkingStatus(status string) error {
	if !slices.Contains(teacherWorkingStatuses, status) {
		return fmt.Errorf("working status must be one of %v", teacherWorkingStatuses)
	}
	return nil
}

// validatePassword checks that the password meets the following requirements:
// - at least 8 characters long
// - at most 72 characters long
// - contains at least one uppercase letter
// - contains at least one lowercase letter
// - contains at least one digit
// - contains at least one special character
func validatePassword(password string) error {
	if len(password) < minPasswordLength {
		return fmt.Errorf("password must be at least %d characters long", minPasswordLength)
	}

	if len(password) > maxPasswordLength {
		return fmt.Errorf("password must be at most %d characters long", maxPasswordLength)
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
		return fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return fmt.Errorf("password must contain at least one digit")
	}
	if !hasSpecial {
		return fmt.Errorf("password must contain at least one special character")
	}

	return nil
}
