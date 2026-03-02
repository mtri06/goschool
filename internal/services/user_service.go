package services

import (
	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/rs/zerolog/log"
)

type UserSvcUserRepo interface {
	UsernameExists(username string) (bool, error)
	Create(user *model.User) error
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
		Username: username,
		Password: hashed,
		Role:     constant.RoleAdmin,
	}

	if err := s.userRepo.Create(admin); err != nil {
		log.Fatal().Err(err).Msgf("failed to create admin user: %s", username)
	}

	log.Info().Msgf("Admin user '%s' created successfully", username)
}
