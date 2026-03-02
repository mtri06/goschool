package repository

import (
	"database/sql"
	"fmt"

	"goschool/pkg/model"
)

type TeacherRepository struct {
	db *sql.DB
}

func NewTeacherRepository(db *sql.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// Create inserts a new user and teacher in a single transaction
func (r *TeacherRepository) Create(user *model.User, teacher *model.Teacher) error {
	// Begin transaction
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Insert user and get the generated ID
	err = tx.
		QueryRow(`
			INSERT INTO users (username, password, email, role)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at, updated_at`,
			user.Username, user.Password, user.Email, user.Role,
		).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	// Set the user ID in the teacher model
	teacher.UserID = user.ID

	// Insert teacher
	err = tx.
		QueryRow(`
			INSERT INTO teachers (user_id, name, subject_id, date_of_birth, gender, hire_date, working_status)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING created_at, updated_at`,
			teacher.UserID, teacher.Name, teacher.SubjectID, teacher.DateOfBirth, teacher.Gender, teacher.HireDate, teacher.WorkingStatus).
		Scan(&teacher.CreatedAt, &teacher.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert teacher: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
