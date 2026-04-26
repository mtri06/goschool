package routes

import (
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/pkg/constant"

	"github.com/go-chi/chi/v5"
)

func MountStudentRoutes(router chi.Router, h *handler.StudentHandler) {
	r := chi.NewRouter()

	r.With(mw.Auth, mw.RequireRole(constant.RoleAdmin)).Post("/", h.CreateStudent)

	router.Mount("/students", r)
}
