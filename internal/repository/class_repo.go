package repository

import "github.com/jmoiron/sqlx"

type ClassRepository struct {
	db *sqlx.DB
}

func NewClassRepository(db *sqlx.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

func (r *ClassRepository) ClassExists(id int64) (bool, error) {
	var exists bool
	err := r.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1)", id)
	return exists, err
}
