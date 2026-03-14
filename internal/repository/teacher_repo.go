package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

	"goschool/pkg/model"

	"github.com/rs/zerolog/log"
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

// List returns a paginated list of teachers and the total count, with optional filters
func (r *TeacherRepository) List(page, pageSize int, name, email string) ([]model.Teacher, int, error) {
	offset := (page - 1) * pageSize

	// Build JOIN and WHERE clause dynamically
	var conditions []string
	var args []any
	join := ""
	if name != "" {
		args = append(args, "%"+name+"%")
		conditions = append(conditions, fmt.Sprintf("t.name ILIKE $%d", len(args)))
	}
	if email != "" {
		join = "JOIN users u ON u.id = t.user_id"
		args = append(args, email+"%")
		conditions = append(conditions, fmt.Sprintf("u.email LIKE $%d", len(args)))
	}
	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var (
		total    int
		teachers []model.Teacher
		countErr error
		listErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		q := fmt.Sprintf(`SELECT COUNT(*) FROM teachers t %s %s`, join, where)
		if err := r.db.QueryRow(q, args...).Scan(&total); err != nil {
			countErr = fmt.Errorf("failed to count teachers: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		// Append LIMIT/OFFSET args after filter args
		listArgs := append(args, pageSize, offset)
		q := fmt.Sprintf(`
			SELECT t.user_id, t.name, t.subject_id, t.date_of_birth, t.gender, t.hire_date, t.working_status, t.created_at, t.updated_at
			FROM teachers t
			%s
			%s
			ORDER BY t.user_id
			LIMIT $%d OFFSET $%d
		`, join, where, len(listArgs)-1, len(listArgs))
		rows, err := r.db.Query(q, listArgs...)
		if err != nil {
			listErr = fmt.Errorf("failed to list teachers: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var t model.Teacher
			if err := rows.Scan(&t.UserID, &t.Name, &t.SubjectID, &t.DateOfBirth, &t.Gender, &t.HireDate, &t.WorkingStatus, &t.CreatedAt, &t.UpdatedAt); err != nil {
				listErr = fmt.Errorf("failed to scan teacher: %w", err)
				return
			}
			teachers = append(teachers, t)
		}
		if err := rows.Err(); err != nil {
			listErr = fmt.Errorf("rows error: %w", err)
		}
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, countErr
	}
	if listErr != nil {
		return nil, 0, listErr
	}

	return teachers, total, nil
}

// Delete removes the teacher and their associated user in a single transaction.
func (r *TeacherRepository) Delete(userID int64) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var res sql.Result
	if res, err = tx.Exec(`DELETE FROM teachers WHERE user_id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete teacher: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		log.Debug().Msg("no rows affected")
		return nil
	}

	if _, err = tx.Exec(`DELETE FROM users WHERE id = $1`, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
