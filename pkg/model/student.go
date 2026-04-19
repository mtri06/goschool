package model

import "time"

type NewStudent struct {
	Username    string    `json:"username" validate:"required"`
	Password    string    `json:"password" validate:"required"`
	Email       *string   `json:"email" validate:"omitempty,email"`
	Name        string    `json:"name" validate:"required"`
	DateOfBirth time.Time `json:"dateOfBirth" validate:"required"`
	Gender      string    `json:"gender" validate:"required"`
}
