package repository

import (
	"database/sql"
	"fmt"

	"goschool/pkg/model"

	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

// EmailExists checks if a user with the given email exists in the database.
// Optional excludeIDs can be passed to exclude specific user IDs from the check.
func (r *UserRepository) EmailExists(email string, excludeIDs ...int64) (bool, error) {
	var exists bool
	var err error
	if len(excludeIDs) > 0 {
		query, args, qErr := sqlx.In("SELECT EXISTS(SELECT 1 FROM users WHERE email = ? AND id NOT IN (?))", email, excludeIDs)
		if qErr != nil {
			return false, fmt.Errorf("failed to build email exists query: %w", qErr)
		}
		query = r.db.Rebind(query)
		err = r.db.Get(&exists, query, args...)
	} else {
		err = r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)", email)
	}
	if err != nil {
		return false, fmt.Errorf("failed to check if email exists: %w", err)
	}
	return exists, nil
}

// UsernameExists checks if a user with the given username exists in the database
func (r *UserRepository) UsernameExists(username string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE username = $1)", username)
	if err != nil {
		return false, fmt.Errorf("failed to check if username exists: %w", err)
	}
	return exists, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(id int64) (*model.User, error) {
	var user model.User
	err := r.db.Get(&user, `SELECT * FROM users WHERE id = $1`, id)
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
	err := r.db.Get(&user, `SELECT * FROM users WHERE username = $1`, username)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

// Create inserts a new user and returns the generated ID
func (r *UserRepository) CreateUser(user *model.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	return r.db.QueryRow(`
		INSERT INTO users (username, password, email, role, name, date_of_birth, gender)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		user.Username, user.Password, user.Email, user.Role, user.Name, user.DateOfBirth, user.Gender,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}
