package service

import (
	"fmt"
	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"slices"
	"strings"
	"time"
	"unicode"

	"github.com/rs/zerolog/log"
)

const (
	minPasswordLength = 8
	maxPasswordLength = 72
)

var (
	allGenders             = []string{constant.GenderMale, constant.GenderFemale, constant.GenderOther}
	teacherWorkingStatuses = []string{
		constant.WorkingStatusActive,
		constant.WorkingStatusInactive,
		constant.WorkingStatusOnLeave,
	}
	allRoles = []string{constant.RoleAdmin, constant.RoleTeacher, constant.RoleStudent}
)

type userSvcUserRepo interface {
	UsernameExists(username string) (bool, error)
	CreateUser(user *model.User) error
	EmailExists(email string, excludeIDs ...int64) (bool, error)
}

type UserService struct {
	userRepo userSvcUserRepo
}

func NewUserService(userRepo userSvcUserRepo) *UserService {
	return &UserService{userRepo: userRepo}
}

// SeedAdminUser creates the admin user if it does not already exist
func (s *UserService) SeedAdminUser() {
	username := env.Env.AdminUsername
	password := env.Env.AdminPassword

	exists, err := s.userRepo.UsernameExists(username)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to check if admin user exists")
	}
	if exists {
		return
	}

	hashed, err := hashPassword(password)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to hash admin password")
	}
	admin := &model.User{
		Username:    username,
		Password:    hashed,
		Role:        constant.RoleAdmin,
		Name:        "admin-user",
		DateOfBirth: time.Now(),
		Gender:      constant.GenderOther,
	}

	if err := s.userRepo.CreateUser(admin); err != nil {
		log.Fatal().Err(err).Msgf("failed to create admin user: %s", username)
	}

	log.Info().Msgf("Admin user '%s' created successfully", username)
}

func (s *UserService) validateUser(user *model.User) error {
	if !slices.Contains(allGenders, user.Gender) {
		return NewError(fmt.Sprintf("gender must be one of %v", allGenders), "invalid_gender", ErrValidationFailed)
	}

	if !slices.Contains(allRoles, user.Role) {
		return NewError(fmt.Sprintf("role must be one of %v", allRoles), "invalid_role", ErrValidationFailed)
	}

	if err := validatePassword(user.Password); err != nil {
		return NewError(fmt.Sprintf("invalid password: %v", err), "invalid_password", ErrValidationFailed)
	}

	exists, err := s.userRepo.UsernameExists(user.Username)
	if err != nil {
		return fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return NewError("username already exists", "username_exists", ErrValidationFailed)
	}

	if user.Email != nil {
		*user.Email = strings.ToLower(*user.Email)
		exists, err := s.userRepo.EmailExists(*user.Email)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return NewError("email already exists", "email_exists", ErrValidationFailed)
		}
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
