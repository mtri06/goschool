package model

type NewSubject struct {
	Name string `json:"name" validate:"required"`
}

type SubjectDetails struct {
	ID     int    `json:"id" db:"id"`
	Name   string `json:"name" db:"name"`
	Status string `json:"status" db:"status"`
}
