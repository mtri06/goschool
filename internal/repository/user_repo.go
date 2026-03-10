package repository

import (
	"database/sql"
	"fmt"

	"goschool/pkg/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// EmailExists checks if a user with the given email exists in the database
func (r *UserRepository) EmailExists(email string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if email exists: %w", err)
	}
	return exists, nil
}

// UsernameExists checks if a user with the given username exists in the database
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if username exists: %w", err)
	}
	return exists, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(`
		SELECT id, username, password, email, role, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by id: %w", err)
	}
	return &user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRow(`
		SELECT id, username, password, email, role, created_at, updated_at
		FROM users WHERE username = $1
	`, username).Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// Create inserts a new user and returns the generated ID
func (r *UserRepository) Create(user *model.User) error {
	err := r.db.
		QueryRow(`
			INSERT INTO users (username, password, email, role)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at`,
			user.Username, user.Password, user.Email, user.Role,
		).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}
