package model

import "time"

type Teacher struct {
	UserID        int64     `json:"user_id" db:"user_id"`
	Name          string    `json:"name" db:"name"`
	SubjectID     int64     `json:"subject_id" db:"subject_id"`
	DateOfBirth   time.Time `json:"date_of_birth" db:"date_of_birth"`
	Gender        string    `json:"gender" db:"gender"`
	HireDate      time.Time `json:"hire_date" db:"hire_date"`
	WorkingStatus string    `json:"working_status" db:"working_status"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time `json:"updated_at" db:"updated_at"`
}

type NewTeacher struct {
	Username      string    `json:"username" validate:"required"`
	Password      string    `json:"password" validate:"required"`
	Email         *string   `json:"email" validate:"omitempty,email"`
	Name          string    `json:"name" validate:"required"`
	SubjectID     int64     `json:"subjectId" validate:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" validate:"required"`
	Gender        string    `json:"gender" validate:"required"`
	HireDate      time.Time `json:"hireDate" validate:"required"`
	WorkingStatus string    `json:"workingStatus" validate:"required"`
}
