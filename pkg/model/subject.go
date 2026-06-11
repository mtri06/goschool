package model

import (
	"time"
)

type NewSubject struct {
	Name string `json:"name" validate:"required"`
}

type ListSubjectsFilter struct {
	Status *string
}
type GetAllSubjectsParams struct {
	Filter  ListSubjectsFilter
	OrderBy OrderBy
}

type SubjectDetails struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
