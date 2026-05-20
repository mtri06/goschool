package repository

import (
	"fmt"

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

// CreateSubject creates a new subject in the database
func (r *SubjectRepository) CreateSubject(newSubject *model.NewSubject) (*model.SubjectDetails, error) {
	subject := &model.SubjectDetails{
		Name:   newSubject.Name,
		Status: constant.SubjectStatusActive,
	}

	err := r.db.QueryRowx(
		`INSERT INTO subjects (name, status) VALUES ($1, $2) 
		 RETURNING id, name, status`,
		newSubject.Name,
		subject.Status,
	).StructScan(subject)

	if err != nil {
		return nil, fmt.Errorf("failed to create subject: %w", err)
	}

	return subject, nil
}
