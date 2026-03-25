package model

import "time"

type User struct {
	ID          int64     `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Password    string    `json:"password" db:"password"`
	Email       *string   `json:"email" db:"email"`
	Role        string    `json:"role" db:"role"`
	Name        string    `json:"name" db:"name"`
	DateOfBirth time.Time `json:"dateOfBirth" db:"date_of_birth"`
	Gender      string    `json:"gender" db:"gender"`
	CreatedAt   time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt   time.Time `json:"updatedAt" db:"updated_at"`
}

type NewTeacher struct {
	Username      string    `json:"username" validate:"required"`
	Password      string    `json:"password" validate:"required"`
	Email         *string   `json:"email" validate:"omitempty,email"`
	Name          string    `json:"name" validate:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" validate:"required"`
	Gender        string    `json:"gender" validate:"required"`
	SubjectID     int64     `json:"subjectId" validate:"required"`
	HireDate      time.Time `json:"hireDate" validate:"required"`
	WorkingStatus string    `json:"workingStatus" validate:"required"`
}

type TeacherDetails struct {
	ID            int64     `json:"id" db:"id"`
	Username      string    `json:"username" db:"username"`
	Email         *string   `json:"email" db:"email"`
	Role          string    `json:"role" db:"role"`
	Name          string    `json:"name" db:"name"`
	DateOfBirth   time.Time `json:"dateOfBirth" db:"date_of_birth"`
	Gender        string    `json:"gender" db:"gender"`
	SubjectID     int64     `json:"subjectId" validate:"required"`
	HireDate      time.Time `json:"hireDate" validate:"required"`
	WorkingStatus string    `json:"workingStatus" validate:"required"`
	CreatedAt     time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt     time.Time `json:"updatedAt" db:"updated_at"`
}
