package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"sync"

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
func (r *UserRepository) CreateUser(user *model.User) error {
	err := r.db.
		QueryRow(`
			INSERT INTO users (username, password, email, role, name, date_of_birth, gender)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, created_at, updated_at`,
			user.Username, user.Password, user.Email, user.Role, user.Name, user.DateOfBirth, user.Gender,
		).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// CreateTeacher inserts a user and a matching user_teacher record in a single transaction
func (r *UserRepository) CreateTeacher(user *model.User, teacher *model.UserTeacher) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	err = tx.QueryRow(`
		INSERT INTO users (username, password, email, role, name, date_of_birth, gender)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		user.Username, user.Password, user.Email, user.Role, user.Name, user.DateOfBirth, user.Gender,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert user: %w", err)
	}

	teacher.UserID = user.ID
	err = tx.QueryRow(`
		INSERT INTO user_teachers (user_id, subject_id, hire_date, working_status)
		VALUES ($1, $2, $3, $4)
		RETURNING created_at, updated_at`,
		teacher.UserID, teacher.SubjectID, teacher.HireDate, teacher.WorkingStatus,
	).Scan(&teacher.CreatedAt, &teacher.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert teacher: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// ListTeachers returns a paginated list of teachers and the total count, with optional filters by name and email
func (r *UserRepository) ListTeachers(page, pageSize int, name, email string) ([]model.TeacherDetails, int, error) {
	offset := (page - 1) * pageSize

	var conditions []string
	var args []any
	conditions = append(conditions, "u.role = 'teacher'")
	if name != "" {
		args = append(args, "%"+name+"%")
		conditions = append(conditions, fmt.Sprintf("u.name ILIKE $%d", len(args)))
	}
	if email != "" {
		email = strings.ToLower(email)
		args = append(args, "%"+email+"%")
		conditions = append(conditions, fmt.Sprintf("u.email LIKE $%d", len(args)))
	}
	where := "WHERE " + strings.Join(conditions, " AND ")

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
		q := fmt.Sprintf(`SELECT COUNT(*) FROM users u %s`, where)
		if err := r.db.QueryRow(q, args...).Scan(&total); err != nil {
			countErr = fmt.Errorf("failed to count teachers: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		listArgs := append(args, pageSize, offset)
		q := fmt.Sprintf(`
			SELECT u.id, u.username, u.email, u.role, u.name, u.date_of_birth, u.gender,
			       t.subject_id, t.hire_date, t.working_status, t.created_at, t.updated_at
			FROM users u
			JOIN user_teachers t ON t.user_id = u.id
			%s
			LIMIT $%d OFFSET $%d
		`, where, len(listArgs)-1, len(listArgs))
		rows, err := r.db.Query(q, listArgs...)
		if err != nil {
			listErr = fmt.Errorf("failed to list teachers: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var t model.TeacherDetails
			if err := rows.Scan(
				&t.ID, &t.Username, &t.Email, &t.Role, &t.Name, &t.DateOfBirth, &t.Gender,
				&t.SubjectID, &t.HireDate, &t.WorkingStatus, &t.CreatedAt, &t.UpdatedAt,
			); err != nil {
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
		return nil, 0, fmt.Errorf("failed to count teachers: %w", countErr)
	}
	if listErr != nil {
		return nil, 0, fmt.Errorf("failed to list teachers: %w", listErr)
	}

	return teachers, total, nil
}

// DeleteTeacher removes the teacher and their associated user in a single transaction.
func (r *UserRepository) DeleteTeacher(teacherID int64) error {
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
	if res, err = tx.Exec(`DELETE FROM user_teachers WHERE user_id = $1`, teacherID); err != nil {
		return fmt.Errorf("failed to delete teacher: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if n == 0 {
		return nil
	}

	if _, err = tx.Exec(`DELETE FROM users WHERE id = $1`, teacherID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
