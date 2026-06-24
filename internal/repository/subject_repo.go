package repository

import (
	"fmt"
	"strings"

	"goschool/pkg/constant"
	"goschool/pkg/model"

	"github.com/jmoiron/sqlx"
)

type SubjectRepository struct {
	db *sqlx.DB
}

func NewSubjectRepository(db *sqlx.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

// Exists checks if a subject with the given ID exists in the database
func (r *SubjectRepository) Exists(id int) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subjects WHERE id = $1)", id)
	if err != nil {
		return false, fmt.Errorf("failed to check if subject exists: %w", err)
	}
	return exists, nil
}

// ExistsByName checks if a subject with the given name exists in the database
func (r *SubjectRepository) ExistsByName(name string) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subjects WHERE name = $1)", name)
	if err != nil {
		return false, fmt.Errorf("failed to check if subject name exists: %w", err)
	}
	return exists, nil
}

// Create creates a new subject in the database
func (r *SubjectRepository) Create(newSubject *model.NewSubject) (*model.SubjectDetails, error) {
	subject := &model.SubjectDetails{
		Name:   newSubject.Name,
		Status: constant.SubjectStatusActive,
	}

	err := r.db.QueryRowx(
		`INSERT INTO subjects (name, status) VALUES ($1, $2) 
		 RETURNING id, name, status, created_at, updated_at`,
		newSubject.Name, subject.Status,
	).StructScan(subject)

	if err != nil {
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}

	return subject, nil
}

// GetAll returns all subjects with optional filtering and ordering
func (r *SubjectRepository) GetAll(params model.GetAllSubjectsParams) ([]model.SubjectDetails, error) {
	query := "SELECT id, name, status, created_at, updated_at FROM subjects"

	var args []any
	if params.Filter.Status != nil {
		query += " WHERE status = $1 "
		args = append(args, *params.Filter.Status)
	}

	// Apply ordering if provided
	query += orderByToSQL(params.OrderBy)

	var subjects []model.SubjectDetails
	err := r.db.Select(&subjects, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list subjects: %w", err)
	}

	return subjects, nil
}

func (r *SubjectRepository) Update(id int, update model.UpdateSubject) error {
	sets := []string{}
	args := []any{}
	argPos := 1

	if update.Name != nil {
		sets = append(sets, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *update.Name)
		argPos++
	}
	if update.Status != nil {
		sets = append(sets, fmt.Sprintf("status = $%d", argPos))
		args = append(args, *update.Status)
		argPos++
	}

	if len(sets) == 0 {
		return nil // Nothing to update
	}

	query := `UPDATE subjects SET ` + strings.Join(sets, ", ") + `, updated_at = NOW() WHERE id = $` + fmt.Sprint(argPos)
	args = append(args, id)
	_, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update subject: %w", err)
	}
	return nil
}
