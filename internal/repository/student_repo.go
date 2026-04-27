package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"sync"

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
		INSERT INTO user_students (user_id, class_id, admission_date)
		VALUES ($1, $2, $3)`,
		userID, newStudent.ClassID, newStudent.AdmissionDate,
	)
	if err != nil {
		return fmt.Errorf("failed to insert student: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetStudentByID retrieves a student with full details by user ID.
func (r *StudentRepository) GetStudentByID(id int64) (*model.StudentDetails, error) {
	var s model.StudentDetails
	var classID *int64
	var className *string
	err := r.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.name, u.date_of_birth, u.gender,
		       s.admission_date, s.graduated, s.graduated_date,
		       c.id, c.name
		FROM users u
		JOIN user_students s ON s.user_id = u.id
		LEFT JOIN classes c ON c.id = s.class_id
		WHERE u.id = $1 AND u.role = $2`, id, constant.RoleStudent,
	).Scan(
		&s.ID, &s.Username, &s.Email, &s.Name, &s.DateOfBirth, &s.Gender,
		&s.AdmissionDate, &s.Graduated, &s.GraduatedDate, &classID, &className,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get student by id: %w", err)
	}
	if classID != nil && className != nil {
		s.Class = &model.StudentDetailsClass{ID: *classID, Name: *className}
	}
	return &s, nil
}

// StudentExists checks if a student with the given user ID exists.
func (r *StudentRepository) StudentExists(id int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = $2)`, id, constant.RoleStudent)
	if err != nil {
		return false, fmt.Errorf("failed to check if student exists: %w", err)
	}
	return exists, nil
}

// UpdateStudent updates user and user_students fields for the given student ID in a single transaction.
func (r *StudentRepository) UpdateStudent(studentID int64, update *model.UpdateStudent) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`UPDATE users SET name = $1, date_of_birth = $2, gender = $3, email = $4, updated_at = NOW() WHERE id = $5 AND role = $6`,
		update.Name, update.DateOfBirth, update.Gender, update.Email, studentID, constant.RoleStudent,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	_, err = tx.Exec(
		`UPDATE user_students SET admission_date = $1, updated_at = NOW() WHERE user_id = $2`,
		update.AdmissionDate, studentID,
	)
	if err != nil {
		return fmt.Errorf("failed to update student: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStudent removes the student and their associated user record.
func (r *StudentRepository) DeleteStudent(studentID int64) error {
	// Delete the user with role = 'student'; cascades to user_students automatically.
	if _, err := r.db.Exec(`DELETE FROM users WHERE id = $1 AND role = $2`, studentID, constant.RoleStudent); err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}
	return nil
}

// ListStudents returns a paginated list of students with optional filters.
// Filters on name/email go into userFilters; filters on class_id/year go into studentFilters.
// When studentFilters are provided, student information is joined.
func (r *StudentRepository) ListStudents(
	p *Pagination, userFilters Filters, studentFilters Filters,
) ([]model.StudentDetails, int, error) {
	limitOffset := p.toLimitOffset()

	userFilters = append(userFilters, NewFilter("role", OpEquals, constant.RoleStudent))
	userFilters.setAlias("u")
	studentFilters.setAlias("s")
	filters := append(userFilters, studentFilters...)
	where, args, err := filters.toWhereClause()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query condition: %w", err)
	}

	var (
		total    int
		students []model.StudentDetails
		countErr error
		listErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		q := "SELECT COUNT(*) FROM users u "
		if len(studentFilters) > 0 {
			q += "JOIN user_students s ON s.user_id = u.id "
		}
		q += where
		if err := r.db.Get(&total, q, args...); err != nil {
			countErr = fmt.Errorf("failed to count students: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		q := fmt.Sprintf(`
			SELECT u.id, u.username, u.email, u.name, u.date_of_birth, u.gender,
			       s.admission_date, s.graduated, s.graduated_date,
			       c.id, c.name
			FROM users u
			JOIN user_students s ON s.user_id = u.id
			LEFT JOIN classes c ON c.id = s.class_id
			%s
			%s
		`, where, limitOffset)
		rows, err := r.db.Query(q, args...)
		if err != nil {
			listErr = fmt.Errorf("failed to list students: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var s model.StudentDetails
			var classID *int64
			var className *string
			if err := rows.Scan(
				&s.ID, &s.Username, &s.Email, &s.Name, &s.DateOfBirth, &s.Gender,
				&s.AdmissionDate, &s.Graduated, &s.GraduatedDate, &classID, &className,
			); err != nil {
				listErr = fmt.Errorf("failed to scan student: %w", err)
				return
			}
			if classID != nil && className != nil {
				s.Class = &model.StudentDetailsClass{ID: *classID, Name: *className}
			}
			students = append(students, s)
		}
		if err := rows.Err(); err != nil {
			listErr = fmt.Errorf("failed to iterate students: %w", err)
		}
	}()

	wg.Wait()

	err = errors.Join(countErr, listErr)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}
