package repository

import (
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/jmoiron/sqlx"
)

type StudentRepository struct {
	db *sqlx.DB
}

func NewStudentRepository(db *sqlx.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) CreateStudent(newStudent *model.NewStudent) error {
	if newStudent == nil {
		return fmt.Errorf("newStudent cannot be nil")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID int64
	err = tx.QueryRow(`
		INSERT INTO users (username, password, email, role, name, date_of_birth, gender)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		newStudent.Username, newStudent.Password, newStudent.Email, constant.RoleStudent,
		newStudent.Name, newStudent.DateOfBirth, newStudent.Gender,
	).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO user_students (user_id, admission_date)
		VALUES ($1, $2)`,
		userID, newStudent.AdmissionDate,
	)
	if err != nil {
		return fmt.Errorf("failed to insert student: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
