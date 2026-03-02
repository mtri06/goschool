package model

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

type AuthTokens struct {
	AccessToken  string
	RefreshToken string
}
