package services

import (
	"fmt"
	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"slices"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

var (
	allGenders             = []string{constant.GenderMale, constant.GenderFemale, constant.GenderOther}
	teacherWorkingStatuses = []string{
		constant.WorkingStatusActive,
		constant.WorkingStatusInactive,
		constant.WorkingStatusOnLeave,
	}
)

type UserSvcUserRepo interface {
	UsernameExists(username string) (bool, error)
	CreateUser(user *model.User) error
	EmailExists(email string) (bool, error)
}

type UserService struct {
	userRepo UserSvcUserRepo
}

func NewUserService(userRepo UserSvcUserRepo) *UserService {
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
		return fmt.Errorf("%w: gender must be one of %v, got: %s", ErrValidationFailed, allGenders, user.Gender)
	}

	if err := validatePassword(user.Password); err != nil {
		return fmt.Errorf("%w: invalid password: %v", ErrValidationFailed, err)
	}

	exists, err := s.userRepo.UsernameExists(user.Username)
	if err != nil {
		return fmt.Errorf("failed to check if username exists: %w", err)
	}
	if exists {
		return fmt.Errorf("%w: username already exists: %s", ErrValidationFailed, user.Username)
	}

	if user.Email != nil {
		*user.Email = strings.ToLower(*user.Email)
		exists, err := s.userRepo.EmailExists(*user.Email)
		if err != nil {
			return fmt.Errorf("failed to check if email exists: %w", err)
		}
		if exists {
			return fmt.Errorf("%w: email already exists: %s", ErrValidationFailed, *user.Email)
		}
	}

	return nil
}
