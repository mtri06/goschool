package routes

import (
	"goschool/internal/api/handler"

	"github.com/go-chi/chi/v5"
)

func MountAuthRoutes(router chi.Router, h *handler.AuthHandler) {
	r := chi.NewRouter()

	r.Post("/login", h.Login)
	r.Post("/logout", h.Logout)

	router.Mount("/auth", r)
}
