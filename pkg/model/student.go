package model

import "time"

type NewStudent struct {
	Username      string    `json:"username" validate:"required"`
	Password      string    `json:"password" validate:"required"`
	Email         *string   `json:"email" validate:"omitempty,email"`
	Name          string    `json:"name" validate:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" validate:"required"`
	Gender        string    `json:"gender" validate:"required"`
	ClassID       *int64    `json:"classId"`
	AdmissionDate time.Time `json:"admissionDate" validate:"required"`
}

type UpdateStudent struct {
	Email         *string   `json:"email" validate:"omitempty,email"`
	Name          string    `json:"name" validate:"required"`
	DateOfBirth   time.Time `json:"dateOfBirth" validate:"required"`
	Gender        string    `json:"gender" validate:"required"`
	AdmissionDate time.Time `json:"admissionDate" validate:"required"`
}

type StudentDetails struct {
	ID            int64      `json:"id" db:"id"`
	Username      string     `json:"username" db:"username"`
	Email         *string    `json:"email" db:"email"`
	Name          string     `json:"name" db:"name"`
	DateOfBirth   time.Time  `json:"dateOfBirth" db:"date_of_birth"`
	Gender        string     `json:"gender" db:"gender"`
	ClassID       *int64     `json:"classId" db:"class_id"`
	AdmissionDate time.Time  `json:"admissionDate" db:"admission_date"`
	Graduated     bool       `json:"graduated" db:"graduated"`
	GraduatedDate *time.Time `json:"graduatedDate" db:"graduated_date"`
}
