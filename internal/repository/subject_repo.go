package repository

import (
	"database/sql"
	"fmt"
)

type SubjectRepository struct {
	db *sql.DB
}

func NewSubjectRepository(db *sql.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

// Exists checks if a subject with the given ID exists in the database
func (r *SubjectRepository) Exists(id int64) (bool, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM subjects WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if subject exists: %w", err)
	}
	return exists, nil
}
