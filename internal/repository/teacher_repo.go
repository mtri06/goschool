package repository

import (
	"database/sql"
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"sync"

	"github.com/jmoiron/sqlx"
)

type TeacherRepository struct {
	db *sqlx.DB
}

func NewTeacherRepository(db *sqlx.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

// CreateTeacher inserts a user and a matching user_teacher record in a single transaction
func (r *TeacherRepository) CreateTeacher(newTeacher *model.NewTeacher) error {
	if newTeacher == nil {
		return fmt.Errorf("newTeacher cannot be nil")
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
		newTeacher.Username, newTeacher.Password, newTeacher.Email, constant.RoleTeacher,
		newTeacher.Name, newTeacher.DateOfBirth, newTeacher.Gender,
	).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO user_teachers (user_id, subject_id, hire_date, working_status)
		VALUES ($1, $2, $3, $4)`,
		userID, newTeacher.SubjectID, newTeacher.HireDate, newTeacher.WorkingStatus,
	)
	if err != nil {
		return fmt.Errorf("failed to insert teacher: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListTeachers returns a paginated list of teachers and the total count, with optional filters by name, email, and working status
func (r *TeacherRepository) ListTeachers(
	p *Pagination, userFilters Filters, teacherFilters Filters,
) ([]model.TeacherDetails, int, error) {
	limitOffset := p.toLimitOffset()

	userFilters.setAlias("u")
	teacherFilters.setAlias("t")
	filters := append(userFilters, teacherFilters...)
	where, args, err := filters.toWhereClause()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to build query condition: %w", err)
	}

	var (
		total    int
		teachers []model.TeacherDetails
		countErr error
		listErr  error
		wg       sync.WaitGroup
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		q := "SELECT COUNT(*) FROM users u "
		if len(teacherFilters) > 0 {
			q += "JOIN user_teachers t ON t.user_id = u.id "
		}
		q += where
		if err := r.db.Get(&total, q, args...); err != nil {
			countErr = fmt.Errorf("failed to count teachers: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		q := fmt.Sprintf(`
			SELECT u.id, u.username, u.email, u.name, u.date_of_birth, u.gender,
			       t.hire_date, t.working_status,
			       s.id, s.name
			FROM users u
			JOIN user_teachers t ON t.user_id = u.id
			LEFT JOIN subjects s ON s.id = t.subject_id
			%s
			%s
		`, where, limitOffset)
		rows, err := r.db.Query(q, args...)
		if err != nil {
			listErr = fmt.Errorf("failed to list teachers: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var t model.TeacherDetails
			var subjectID *int64
			var subjectName *string

			if err := rows.Scan(
				&t.ID, &t.Username, &t.Email, &t.Name, &t.DateOfBirth, &t.Gender,
				&t.HireDate, &t.WorkingStatus, &subjectID, &subjectName,
			); err != nil {
				listErr = fmt.Errorf("failed to scan teacher: %w", err)
				return
			}
			if subjectID != nil && subjectName != nil {
				t.Subject = &model.TeacherDetailsSubject{ID: *subjectID, Name: *subjectName}
			}
			teachers = append(teachers, t)
		}
		if err := rows.Err(); err != nil {
			listErr = fmt.Errorf("failed to iterate teachers: %w", err)
		}
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, fmt.Errorf("failed to count teachers: %w", countErr)
	}
	if listErr != nil {
		return nil, 0, fmt.Errorf("failed to list teachers: %w", listErr)
	}

	return teachers, total, nil
}

// GetTeacherByID retrieves a teacher with full details by user ID.
func (r *TeacherRepository) GetTeacherByID(id int64) (*model.TeacherDetails, error) {
	var t model.TeacherDetails
	var subjectID *int64
	var subjectName *string
	err := r.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.name, u.date_of_birth, u.gender,
		       t.hire_date, t.working_status,
		       s.id, s.name
		FROM users u
		JOIN user_teachers t ON t.user_id = u.id
		LEFT JOIN subjects s ON s.id = t.subject_id
		WHERE u.id = $1 AND u.role = $2`, id, constant.RoleTeacher,
	).Scan(
		&t.ID, &t.Username, &t.Email, &t.Name, &t.DateOfBirth, &t.Gender,
		&t.HireDate, &t.WorkingStatus, &subjectID, &subjectName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get teacher by id: %w", err)
	}
	if subjectID != nil && subjectName != nil {
		t.Subject = &model.TeacherDetailsSubject{ID: *subjectID, Name: *subjectName}
	}
	return &t, nil
}

// TeacherExists checks if a teacher with the given user ID exists.
func (r *TeacherRepository) TeacherExists(id int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = 'teacher')`, id)
	if err != nil {
		return false, fmt.Errorf("failed to check if teacher exists: %w", err)
	}
	return exists, nil
}

// UpdateTeacher updates user and user_teacher fields for the given teacher ID in a single transaction.
func (r *TeacherRepository) UpdateTeacher(teacherID int64, update *model.UpdateTeacher) error {
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Update users with role = 'teacher' condition to verify the ID belongs to a teacher
	_, err = tx.Exec(
		`UPDATE users SET name = $1, date_of_birth = $2, gender = $3, email = $4, updated_at = NOW() WHERE id = $5 AND role = $6`,
		update.Name, update.DateOfBirth, update.Gender, update.Email, teacherID, constant.RoleTeacher,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if _, err := tx.Exec(
		`UPDATE user_teachers SET subject_id = $1, hire_date = $2, working_status = $3, updated_at = NOW() WHERE user_id = $4`,
		update.SubjectID, update.HireDate, update.WorkingStatus, teacherID,
	); err != nil {
		return fmt.Errorf("failed to update teacher: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteTeacher removes the teacher and their associated user in a single transaction.
func (r *TeacherRepository) DeleteTeacher(teacherID int64) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete the user with role = 'teacher'; cascades to user_teachers automatically.
	if _, err = tx.Exec(`DELETE FROM users WHERE id = $1 AND role = $2`, teacherID, constant.RoleTeacher); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Explicit delete of the teacher record (no-op when cascade already removed it).
	if _, err = tx.Exec(`DELETE FROM user_teachers WHERE user_id = $1`, teacherID); err != nil {
		return fmt.Errorf("failed to delete teacher: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
