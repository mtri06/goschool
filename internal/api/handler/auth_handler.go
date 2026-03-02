package handler

import (
	"net/http"
	"time"

	"goschool/internal/env"
	"goschool/pkg/constant"
	"goschool/pkg/httpx"
	"goschool/pkg/model"
)

type AuthService interface {
	Login(username, password string) (*model.AuthTokens, error)
}

type AuthHandler struct {
	authSvc AuthService
	errMap  httpx.APIErrorMap
}

func NewAuthHandler(authSvc AuthService, errMap httpx.APIErrorMap) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, errMap: errMap}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	req, err := httpx.DecodeBody[model.LoginRequest](r)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	tokens, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	// Set access token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constant.CookieAccessToken,
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   env.Env.JWTAccessExpiresMins * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	// Set refresh token cookie
	http.SetCookie(w, &http.Cookie{
		Name:     constant.CookieRefreshToken,
		Value:    tokens.RefreshToken,
		Path:     "/",
		MaxAge:   env.Env.JWTRefreshExpiresDays * 24 * int(time.Hour.Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}
