package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"goschool/pkg/constant"
	"goschool/pkg/model"
	"strings"
	"sync"
	"time"

	"github.com/jmoiron/sqlx"
)

type StudentRepository struct {
	db *sqlx.DB
}

func NewStudentRepository(db *sqlx.DB) *StudentRepository {
	return &StudentRepository{db: db}
}

func (r *StudentRepository) Create(newStudent *model.NewStudent) (*model.StudentDetails, error) {
	if newStudent == nil {
		return nil, fmt.Errorf("newStudent cannot be nil")
	}

	tx, err := r.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var userID int
	var updatedAt, createdAt time.Time
	err = tx.QueryRow(`
		INSERT INTO users (username, password, email, role, name, date_of_birth, gender)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at`,
		newStudent.Username, newStudent.Password, newStudent.Email, constant.RoleStudent,
		newStudent.Name, newStudent.DateOfBirth, newStudent.Gender,
	).Scan(&userID, &createdAt, &updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	var graduated bool
	var graduatedDate *time.Time
	err = tx.QueryRow(`
		INSERT INTO user_students (user_id, class_id, admission_date)
		VALUES ($1, $2, $3)
		RETURNING graduated, graduated_date`,
		userID, newStudent.ClassID, newStudent.AdmissionDate,
	).Scan(&graduated, &graduatedDate)
	if err != nil {
		return nil, fmt.Errorf("failed to insert student: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	details := &model.StudentDetails{
		ID:            userID,
		Username:      newStudent.Username,
		Email:         newStudent.Email,
		Name:          newStudent.Name,
		DateOfBirth:   newStudent.DateOfBirth,
		Gender:        newStudent.Gender,
		ClassID:       newStudent.ClassID,
		AdmissionDate: newStudent.AdmissionDate,
		Graduated:     graduated,
		GraduatedDate: graduatedDate,
		CreatedAt:     createdAt,
		UpdatedAt:     updatedAt,
	}

	return details, nil
}

// GetDetailsByID retrieves a student with full details by user ID.
func (r *StudentRepository) GetDetailsByID(id int) (*model.StudentDetails, error) {
	var s model.StudentDetails
	var className *string
	err := r.db.QueryRow(`
		SELECT u.id, u.username, u.email, u.name, u.date_of_birth, u.gender,
		       s.admission_date, s.graduated, s.graduated_date, s.class_id,
		       c.name, u.created_at, u.updated_at
		FROM users u
		JOIN user_students s ON s.user_id = u.id
		LEFT JOIN classes c ON c.id = s.class_id
		WHERE u.id = $1 AND u.role = $2`, id, constant.RoleStudent,
	).Scan(
		&s.ID, &s.Username, &s.Email, &s.Name, &s.DateOfBirth, &s.Gender,
		&s.AdmissionDate, &s.Graduated, &s.GraduatedDate, &s.ClassID,
		&className, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get student by id: %w", err)
	}
	if s.ClassID != nil && className != nil {
		s.Class = &model.StudentDetailsClass{ID: *s.ClassID, Name: *className}
	}
	return &s, nil
}

// Exists checks if a student with the given user ID exists.
func (r *StudentRepository) Exists(id int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND role = $2)`, id, constant.RoleStudent)
	if err != nil {
		return false, fmt.Errorf("failed to check if student exists: %w", err)
	}
	return exists, nil
}

// Update updates user and user_students fields for the given student ID in a single transaction.
func (r *StudentRepository) Update(studentID int, update *model.UpdateStudent) error {
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

// Delete removes the student user and associated user_students record.
func (r *StudentRepository) Delete(studentID int) error {
	// Delete the user with role = 'student'; cascades to user_students automatically.
	if _, err := r.db.Exec(`DELETE FROM users WHERE id = $1 AND role = $2`, studentID, constant.RoleStudent); err != nil {
		return fmt.Errorf("failed to delete student: %w", err)
	}
	return nil
}

var studentOrderByMap = map[string]string{
	"id":         "u.id",
	"name":       "u.name",
	"class_id":   "s.class_id",
	"created_at": "u.created_at",
	"updated_at": "u.updated_at",
}

// List returns a paginated list of students with optional filters.
// Filters on name/email go into userFilters; filters on class_id/year go into studentFilters.
// When studentFilters are provided, student information is joined.
func (r *StudentRepository) List(params model.ListStudentsParams) ([]model.StudentDetails, int, error) {
	var args []any
	var whereClauses []string

	whereClauses = append(whereClauses, "u.role = $1")
	args = append(args, constant.RoleStudent)
	if params.Filter.Name != nil {
		whereClauses = append(whereClauses, "u.name ILIKE $"+fmt.Sprint(len(args)+1))
		args = append(args, "%"+*params.Filter.Name+"%")
	}
	if params.Filter.Email != nil {
		email := strings.ToLower(*params.Filter.Email)
		whereClauses = append(whereClauses, "u.email ILIKE $"+fmt.Sprint(len(args)+1))
		args = append(args, "%"+email+"%")
	}
	if params.Filter.ClassID != nil {
		whereClauses = append(whereClauses, "s.class_id = $"+fmt.Sprint(len(args)+1))
		args = append(args, *params.Filter.ClassID)
	}
	if params.Filter.Graduated != nil {
		whereClauses = append(whereClauses, "s.graduated = $"+fmt.Sprint(len(args)+1))
		args = append(args, *params.Filter.Graduated)
	}

	where := "WHERE " + strings.Join(whereClauses, " AND ")

	for i := range params.OrderBy {
		mField, ok := studentOrderByMap[params.OrderBy[i].Field]
		if !ok {
			return nil, 0, fmt.Errorf("invalid order by field: %s", params.OrderBy[i].Field)
		}
		params.OrderBy[i].Field = mField
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
		q := `SELECT COUNT(*) FROM users u 
			JOIN user_students s ON s.user_id = u.id
			` + where
		if err := r.db.Get(&total, q, args...); err != nil {
			countErr = fmt.Errorf("failed to count students: %w", err)
		}
	}()

	go func() {
		defer wg.Done()
		q := fmt.Sprintf(`
			SELECT u.id, u.username, u.email, u.name, u.date_of_birth, u.gender,
			       s.admission_date, s.graduated, s.graduated_date, s.class_id,
			       c.name, u.created_at, u.updated_at
			FROM users u
			JOIN user_students s ON s.user_id = u.id
			LEFT JOIN classes c ON c.id = s.class_id
			%s %s %s
		`, where, orderByToSQL(params.OrderBy), paginationToSQL(params.Pagin))
		rows, err := r.db.Query(q, args...)
		if err != nil {
			listErr = fmt.Errorf("failed to list students: %w", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var s model.StudentDetails
			var className *string
			if err := rows.Scan(
				&s.ID, &s.Username, &s.Email, &s.Name, &s.DateOfBirth, &s.Gender,
				&s.AdmissionDate, &s.Graduated, &s.GraduatedDate, &s.ClassID,
				&className, &s.CreatedAt, &s.UpdatedAt,
			); err != nil {
				listErr = fmt.Errorf("failed to scan student: %w", err)
				return
			}
			if s.ClassID != nil && className != nil {
				s.Class = &model.StudentDetailsClass{ID: *s.ClassID, Name: *className}
			}
			students = append(students, s)
		}
		if err := rows.Err(); err != nil {
			listErr = fmt.Errorf("failed to iterate students: %w", err)
		}
	}()

	wg.Wait()

	err := errors.Join(countErr, listErr)
	if err != nil {
		return nil, 0, err
	}

	return students, total, nil
}
