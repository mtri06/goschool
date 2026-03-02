package services

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthSvcUserRepo interface {
	GetByUsername(username string) (*model.User, error)
}

type AuthSvcTokenRepo interface {
	CreateToken(token *model.Token) error
	RevokeByBody(body string) error
}

type AuthService struct {
	userRepo  AuthSvcUserRepo
	tokenRepo AuthSvcTokenRepo
}

func NewAuthService(userRepo AuthSvcUserRepo, tokenRepo AuthSvcTokenRepo) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
	}
}

// Login validates credentials and returns access + refresh tokens
func (s *AuthService) Login(username, password string) (*model.AuthTokens, error) {
	// Find user by username
	user, err := s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate access token (JWT)
	accessToken, err := generateAccessToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token (opaque random string)
	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token in DB
	token := &model.Token{
		Body:      refreshToken,
		UserID:    user.ID,
		Type:      constant.TokenTypeRefresh,
		ExpiresAt: time.Now().Add(time.Duration(env.Env.JWTRefreshExpiresDays) * 24 * time.Hour),
	}
	if err := s.tokenRepo.CreateToken(token); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.AuthTokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Logout revokes the given refresh token in the DB
func (s *AuthService) Logout(refreshToken string) error {
	if err := s.tokenRepo.RevokeByBody(refreshToken); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	return nil
}

func generateAccessToken(user *model.User) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"iat":  now.Unix(),
		"exp":  now.Add(time.Duration(env.Env.JWTAccessExpiresMins) * time.Minute).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(env.Env.JWTSecret))
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// hashPassword hashes a plaintext password using bcrypt
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}
