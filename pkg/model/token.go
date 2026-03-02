package model

import "time"

type Token struct {
	ID            int64     `json:"id" db:"id"`
	Body          string    `json:"body" db:"body"`
	UserID        int64     `json:"user_id" db:"user_id"`
	Type          string    `json:"type" db:"type"`
	ExpiresAt     time.Time `json:"expires_at" db:"expires_at"`
	IsRevoked     bool      `json:"is_revoked" db:"is_revoked"`
	IsUsed        bool      `json:"is_used" db:"is_used"`
	IsBlacklisted bool      `json:"is_blacklisted" db:"is_blacklisted"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
