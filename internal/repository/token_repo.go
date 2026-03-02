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
