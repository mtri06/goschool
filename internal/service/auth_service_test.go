package service

import (
	"errors"
	"os"
	"testing"
	"time"

	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "test-secret-key-for-unit-tests"

// ---------------------------------------------------------------------------
// TestMain — seed env.Env so generateAccessToken works without a .env file
// ---------------------------------------------------------------------------

func TestMain(m *testing.M) {
	env.Env.JWTSecret = testJWTSecret
	env.Env.JWTAccessExpiresMins = 15
	env.Env.JWTRefreshExpiresDays = 7
	os.Exit(m.Run())
}

// ---------------------------------------------------------------------------
// Mocks
// ---------------------------------------------------------------------------

type mockUserRepo struct {
	getByUsernameFn func(username string) (*model.User, error)
	getByIDFn       func(id int64) (*model.User, error)
}

func (m *mockUserRepo) GetByUsername(username string) (*model.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(username)
	}
	return nil, nil
}
func (m *mockUserRepo) GetByID(id int64) (*model.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(id)
	}
	return nil, nil
}

type mockTokenRepo struct {
	createTokenFn   func(token *model.Token) error
	revokeByBodyFn  func(body string) error
	getByBodyFn     func(body string) (*model.Token, error)
	markUsedFn      func(id int64) error
	blacklistUserFn func(userID int64) error
}

func (m *mockTokenRepo) CreateToken(token *model.Token) error {
	if m.createTokenFn != nil {
		return m.createTokenFn(token)
	}
	return nil
}
func (m *mockTokenRepo) RevokeByBody(body string) error {
	if m.revokeByBodyFn != nil {
		return m.revokeByBodyFn(body)
	}
	return nil
}
func (m *mockTokenRepo) GetByBody(body string) (*model.Token, error) {
	if m.getByBodyFn != nil {
		return m.getByBodyFn(body)
	}
	return nil, nil
}
func (m *mockTokenRepo) MarkUsed(id int64) error {
	if m.markUsedFn != nil {
		return m.markUsedFn(id)
	}
	return nil
}
func (m *mockTokenRepo) BlacklistUserTokens(userID int64) error {
	if m.blacklistUserFn != nil {
		return m.blacklistUserFn(userID)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// makeAccessToken signs a JWT with the test secret and the given expiry.
// Pass exp in the past to create an already-expired token.
func makeAccessToken(userID int64, role string, exp time.Time) string {
	claims := jwt.MapClaims{
		"sub":  float64(userID),
		"role": role,
		"iat":  time.Now().Unix(),
		"exp":  exp.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(testJWTSecret))
	return signed
}

func validRefreshTokenDoc(userID int64) *model.Token {
	return &model.Token{
		ID:            1,
		Body:          "some-opaque-refresh-token",
		UserID:        userID,
		Type:          constant.TokenTypeRefresh,
		ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
		IsRevoked:     false,
		IsUsed:        false,
		IsBlacklisted: false,
	}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

func TestAuthService_RefreshTokens(t *testing.T) {
	const userID int64 = 42
	const refreshBody = "some-opaque-refresh-token"

	expiredAT := makeAccessToken(userID, constant.RoleAdmin, time.Now().Add(-1*time.Hour))
	validAT := makeAccessToken(userID, constant.RoleAdmin, time.Now().Add(15*time.Minute))

	anyErr := errors.New("any error")
	tests := []struct {
		name                string
		accessToken         string
		refreshToken        string
		userRepo            *mockUserRepo
		tokenRepo           *mockTokenRepo
		wantErr             error
		wantBlacklistCalled bool
	}{
		{
			name:         "success — expired access token and valid refresh token",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo: &mockUserRepo{
				getByIDFn: func(id int64) (*model.User, error) {
					return &model.User{ID: id, Role: constant.RoleAdmin}, nil
				},
			},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) { return validRefreshTokenDoc(userID), nil },
			},
			wantErr: nil,
		},
		{
			name:         "access token not yet expired",
			accessToken:  validAT,
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo:    &mockTokenRepo{},
			wantErr:      ErrUnauthorized,
		},
		{
			name:         "access token is an invalid string",
			accessToken:  "not.a.jwt",
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo:    &mockTokenRepo{},
			wantErr:      ErrUnauthorized,
		},
		{
			name:         "refresh token not found in DB",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) { return nil, nil },
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:         "refresh token is expired",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) {
					doc := validRefreshTokenDoc(userID)
					doc.ExpiresAt = time.Now().Add(-1 * time.Hour)
					return doc, nil
				},
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:         "refresh token belongs to a different user",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) {
					return validRefreshTokenDoc(999), nil // wrong user
				},
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:         "refresh token is blacklisted",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) {
					doc := validRefreshTokenDoc(userID)
					doc.IsBlacklisted = true
					return doc, nil
				},
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:                "refresh token already used — triggers full blacklist",
			accessToken:         expiredAT,
			refreshToken:        refreshBody,
			userRepo:            &mockUserRepo{},
			wantBlacklistCalled: true,
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) {
					doc := validRefreshTokenDoc(userID)
					doc.IsUsed = true
					return doc, nil
				},
				blacklistUserFn: func(uid int64) error { return nil },
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:                "refresh token revoked — triggers full blacklist",
			accessToken:         expiredAT,
			refreshToken:        refreshBody,
			userRepo:            &mockUserRepo{},
			wantBlacklistCalled: true,
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) {
					doc := validRefreshTokenDoc(userID)
					doc.IsRevoked = true
					return doc, nil
				},
				blacklistUserFn: func(uid int64) error { return nil },
			},
			wantErr: ErrUnauthorized,
		},
		{
			name:         "user no longer exists in DB",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo: &mockUserRepo{
				getByIDFn: func(id int64) (*model.User, error) { return nil, nil }, // user not found
			},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) { return validRefreshTokenDoc(userID), nil },
				markUsedFn:  func(id int64) error { return nil },
			},
			wantErr: ErrNotFound,
		},
		{
			name:         "DB error getting refresh token",
			accessToken:  expiredAT,
			refreshToken: refreshBody,
			userRepo:     &mockUserRepo{},
			tokenRepo: &mockTokenRepo{
				getByBodyFn: func(body string) (*model.Token, error) {
					return nil, errors.New("connection refused")
				},
			},
			wantErr: anyErr, // just check err != nil
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blacklistCalled := false
			if tc.wantBlacklistCalled {
				orig := tc.tokenRepo.blacklistUserFn
				tc.tokenRepo.blacklistUserFn = func(uid int64) error {
					blacklistCalled = true
					if orig != nil {
						return orig(uid)
					}
					return nil
				}
			}

			svc := NewAuthService(tc.userRepo, tc.tokenRepo)
			tokens, err := svc.RefreshTokens(tc.accessToken, tc.refreshToken)

			if tc.wantErr != nil {
				// If wantErr is anyErr, we just check that an error occurred, without asserting on the exact error message.
				if tc.wantErr == anyErr {
					if err == nil {
						t.Fatalf("expected an error but got nil")
					}
				} else {
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf("expected error %v, got %v", tc.wantErr, err)
					}
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if tokens == nil {
					t.Fatal("expected tokens to be returned, got nil")
				}
				if tokens.AccessToken == "" {
					t.Error("access token should not be empty")
				}
				if tokens.RefreshToken == "" {
					t.Error("refresh token should not be empty")
				}
				if tokens.AccessToken == tc.accessToken {
					t.Error("new access token should differ from the original")
				}
				if tokens.RefreshToken == tc.refreshToken {
					t.Error("new refresh token should differ from the original")
				}

				// Validate the returned access token is a legitimate, non-expired JWT
				parsed, parseErr := jwt.Parse(tokens.AccessToken, func(t *jwt.Token) (any, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, errors.New("unexpected signing method")
					}
					return []byte(testJWTSecret), nil
				})
				if parseErr != nil {
					t.Fatalf("returned access token failed to parse: %v", parseErr)
				}
				if !parsed.Valid {
					t.Error("returned access token is not valid")
				}
				claims, ok := parsed.Claims.(jwt.MapClaims)
				if !ok {
					t.Fatal("returned access token has unexpected claims type")
				}
				expTime, _ := claims.GetExpirationTime()
				if expTime == nil || expTime.Before(time.Now()) {
					t.Error("returned access token should not be expired")
				}
				if claims["sub"] == nil {
					t.Error("returned access token missing sub claim")
				}
				if claims["role"] == nil {
					t.Error("returned access token missing role claim")
				}
			}

			if tc.wantBlacklistCalled && !blacklistCalled {
				t.Error("expected BlacklistUserTokens to be called but it was not")
			}
		})
	}
}
