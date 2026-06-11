package model

import (
	"time"
)

type NewTeacher struct {
	Username      string    `json:"username" validate:"required"`
	Password      string    `json:"password" validate:"required"`
	Email         *string   `json:"email" validate:"omitempty,email"`
	Name          string    `json:"name" validate:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" validate:"required"`
	Gender        string    `json:"gender" validate:"required"`
	SubjectID     *int      `json:"subjectId"`
	HireDate      time.Time `json:"hireDate" validate:"required"`
	WorkingStatus string    `json:"workingStatus" validate:"required"`
}

type UpdateTeacher struct {
	Email         *string   `json:"email" validate:"omitempty,email"`
	Name          string    `json:"name" validate:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" validate:"required"`
	Gender        string    `json:"gender" validate:"required"`
	SubjectID     *int      `json:"subjectId"`
	HireDate      time.Time `json:"hireDate" validate:"required"`
	WorkingStatus string    `json:"workingStatus" validate:"required"`
}

type ListTeacherFilter struct {
	Name          *string
	Email         *string
	WorkingStatus *string
	SubjectID     *int
}
type ListTeachersParams struct {
	Filter  ListTeacherFilter
	OrderBy OrderBy
	Pagin   Pagination
}

type TeacherDetailsSubject struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
type TeacherDetails struct {
	ID            int                    `json:"id" db:"id"`
	Username      string                 `json:"username" db:"username"`
	Email         *string                `json:"email" db:"email"`
	Name          string                 `json:"name" db:"name"`
	DateOfBirth   time.Time              `json:"dateOfBirth" db:"date_of_birth"`
	Gender        string                 `json:"gender" db:"gender"`
	SubjectID     *int                   `json:"subjectId" db:"subject_id"`
	Subject       *TeacherDetailsSubject `json:"subject"`
	HireDate      time.Time              `json:"hireDate" db:"hire_date"`
	WorkingStatus string                 `json:"workingStatus" db:"working_status"`
	CreatedAt     time.Time              `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time              `json:"updatedAt" db:"updated_at"`
}
