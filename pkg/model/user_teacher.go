package model

import "time"

type UserTeacher struct {
	UserID        int64     `json:"user_id" db:"user_id"`
	SubjectID     int64     `json:"subject_id" db:"subject_id"`
	HireDate      time.Time `json:"hire_date" db:"hire_date"`
	WorkingStatus string    `json:"working_status" db:"working_status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}
