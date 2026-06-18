package service

import (
	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"time"

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

type userRepo interface {
	UsernameExists(username string) (bool, error)
	CreateUser(user *model.User) error
	EmailExists(email string) (bool, error)
}

type UserService struct {
	userRepo userRepo
}

func NewUserService(userRepo userRepo) *UserService {
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
