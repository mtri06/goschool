package service

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type authSvcUserRepo interface {
	GetByUsername(username string) (*model.User, error)
	GetByID(id int64) (*model.User, error)
}

type authSvcTokenRepo interface {
	CreateToken(token *model.Token) error
	RevokeByBody(body string) error
	GetByBody(body string) (*model.Token, error)
	MarkUsed(id int64) error
	BlacklistUserTokens(userID int64) error
}

type AuthService struct {
	userRepo  authSvcUserRepo
	tokenRepo authSvcTokenRepo
}

func NewAuthService(userRepo authSvcUserRepo, tokenRepo authSvcTokenRepo) *AuthService {
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
		return nil, NewError("wrong credentials", "wrong_credentials", ErrUnauthorized)
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, NewError("wrong credentials", "wrong_credentials", ErrUnauthorized)
	}

	// Generate access token (JWT)
	accessToken, err := generateAccessToken(user.ID, user.Role)
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

// RefreshTokens validates the existing token pair and issues a new one
func (s *AuthService) RefreshTokens(accessToken, refreshToken string) (*model.AuthTokens, error) {
	// Parse access token — allow expired tokens
	aToken, err := jwt.Parse(accessToken, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrUnauthorized
		}
		return []byte(env.Env.JWTSecret), nil
	})
	if err != nil && !errors.Is(err, jwt.ErrTokenExpired) {
		log.Warn().Err(err).Msg("Failed to parse access token")
		return nil, ErrUnauthorized
	}

	aClaims, ok := aToken.Claims.(jwt.MapClaims)
	if !ok {
		log.Warn().Msg("Failed to parse access token claims")
		return nil, ErrUnauthorized
	}

	// Access token must be expired
	expTime, _ := aClaims.GetExpirationTime()
	if expTime != nil && expTime.Unix() > time.Now().Unix() {
		log.Warn().Msg("Access token has not expired")
		return nil, ErrUnauthorized
	}

	// Get user ID from access token claims
	subFloat, ok := aClaims["sub"].(float64)
	if !ok {
		log.Warn().Msg("Access token 'sub' claim is missing or invalid")
		return nil, ErrUnauthorized
	}
	userID := int64(subFloat)

	// Look up refresh token in DB
	rToken, err := s.tokenRepo.GetByBody(refreshToken)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get refresh token")
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}
	if rToken == nil {
		log.Warn().Msg("Refresh token not found")
		return nil, ErrUnauthorized
	}

	// Refresh token must not be expired
	if rToken.ExpiresAt.Before(time.Now()) {
		log.Warn().Msg("Refresh token is expired")
		return nil, ErrUnauthorized
	}

	// Both tokens must belong to the same user
	if rToken.UserID != userID {
		log.Warn().
			Int64("token_user_id", rToken.UserID).Int64("claims_user_id", userID).
			Msg("Refresh token user ID does not match access token claims")
		return nil, ErrUnauthorized
	}

	// If blacklisted — hard reject
	if rToken.IsBlacklisted {
		log.Warn().Int64("user_id", userID).Msg("Attempt to use blacklisted refresh token")
		return nil, ErrUnauthorized
	}

	// If already used or revoked — token reuse attack: blacklist all user tokens
	if rToken.IsUsed || rToken.IsRevoked {
		log.Warn().Int64("user_id", userID).Msg("Attempt to reuse or use revoked refresh token")
		_ = s.tokenRepo.BlacklistUserTokens(userID)
		return nil, ErrUnauthorized
	}

	// Mark this refresh token as used
	if err := s.tokenRepo.MarkUsed(rToken.ID); err != nil {
		return nil, fmt.Errorf("failed to mark token as used: %w", err)
	}

	// Get full user record (needed for role in new JWT)
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		log.Warn().Int64("user_id", userID).Msg("User not found")
		return nil, ErrNotFound
	}

	// Generate new access + refresh tokens
	newAccessToken, err := generateAccessToken(user.ID, user.Role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	newRToken := &model.Token{
		Body:      newRefreshToken,
		UserID:    user.ID,
		Type:      constant.TokenTypeRefresh,
		ExpiresAt: time.Now().Add(time.Duration(env.Env.JWTRefreshExpiresDays) * 24 * time.Hour),
	}
	if err := s.tokenRepo.CreateToken(newRToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &model.AuthTokens{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func generateAccessToken(userID int64, userRole string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  userID,
		"role": userRole,
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
