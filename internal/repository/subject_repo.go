package repository

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SubjectRepository struct {
	db *sqlx.DB
}

func NewSubjectRepository(db *sqlx.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

// Exists checks if a subject with the given ID exists in the database
func (r *SubjectRepository) Exists(id int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM subjects WHERE id = $1)", id)
	if err != nil {
		return false, fmt.Errorf("failed to check if subject exists: %w", err)
	}
	return exists, nil
}
