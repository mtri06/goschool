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
