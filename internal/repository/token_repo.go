package repository

import (
	"database/sql"
	"fmt"

	"goschool/pkg/model"
)

type TokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// GetByBody retrieves a token by its body
func (r *TokenRepository) GetByBody(body string) (*model.Token, error) {
	var t model.Token
	err := r.db.QueryRow(`
		SELECT id, body, user_id, type, expires_at, is_revoked, is_used, is_blacklisted, created_at, updated_at
		FROM tokens WHERE body = $1
	`, body).Scan(&t.ID, &t.Body, &t.UserID, &t.Type, &t.ExpiresAt, &t.IsRevoked, &t.IsUsed, &t.IsBlacklisted, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get token by body: %w", err)
	}
	return &t, nil
}

// MarkUsed sets is_used = true for the token with the given ID
func (r *TokenRepository) MarkUsed(id int64) error {
	_, err := r.db.Exec(`UPDATE tokens SET is_used = true WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to mark token as used: %w", err)
	}
	return nil
}

// BlacklistUserTokens sets is_blacklisted = true for all active refresh tokens of a user
func (r *TokenRepository) BlacklistUserTokens(userID int64) error {
	_, err := r.db.Exec(`
		UPDATE tokens SET is_blacklisted = true
		WHERE user_id = $1 AND type = 'refresh_token' AND is_blacklisted = false
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to blacklist user tokens: %w", err)
	}
	return nil
}

// RevokeByBody sets is_revoked = true for the token matching the given body
func (r *TokenRepository) RevokeByBody(body string) error {
	_, err := r.db.Exec(`UPDATE tokens SET is_revoked = true WHERE body = $1`, body)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

// CreateToken inserts a new token and returns the generated ID
func (r *TokenRepository) CreateToken(token *model.Token) error {
	err := r.db.QueryRow(`
		INSERT INTO tokens (body, user_id, type, expires_at, is_revoked, is_used, is_blacklisted)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`, token.Body, token.UserID, token.Type, token.ExpiresAt, token.IsRevoked, token.IsUsed, token.IsBlacklisted).Scan(&token.ID, &token.CreatedAt, &token.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert token: %w", err)
	}
	return nil
}
