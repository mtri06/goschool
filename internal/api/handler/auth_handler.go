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
	Logout(refreshToken string) error
	RefreshTokens(accessToken, refreshToken string) (*model.AuthTokens, error)
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
		MaxAge:   env.Env.JWTRefreshExpiresDays * 24 * int(time.Hour.Seconds()), // Match refresh token cookies expiry
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

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(constant.CookieRefreshToken)
	if err != nil || cookie.Value == "" {
		httpx.RenderError(w, r, h.errMap, httpx.ErrUnauthorized)
		return
	}

	if err := h.authSvc.Logout(cookie.Value); err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	// Clear both cookies
	http.SetCookie(w, &http.Cookie{
		Name:     constant.CookieAccessToken,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     constant.CookieRefreshToken,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	accessCookie, err := r.Cookie(constant.CookieAccessToken)
	if err != nil || accessCookie.Value == "" {
		httpx.RenderError(w, r, h.errMap, httpx.ErrUnauthorized)
		return
	}

	refreshCookie, err := r.Cookie(constant.CookieRefreshToken)
	if err != nil || refreshCookie.Value == "" {
		httpx.RenderError(w, r, h.errMap, httpx.ErrUnauthorized)
		return
	}

	tokens, err := h.authSvc.RefreshTokens(accessCookie.Value, refreshCookie.Value)
	if err != nil {
		httpx.RenderError(w, r, h.errMap, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     constant.CookieAccessToken,
		Value:    tokens.AccessToken,
		Path:     "/",
		MaxAge:   env.Env.JWTRefreshExpiresDays * 24 * int(time.Hour.Seconds()), // Match refresh token cookies expiry
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
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
