package routes

import (
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/pkg/constant"

	"github.com/go-chi/chi/v5"
)

func MountTeacherRoutes(router chi.Router, h *handler.TeacherHandler) {
	r := chi.NewRouter()

	r.With(mw.Auth, mw.RequireRole(constant.RoleAdmin)).Post("/", h.CreateTeacher)
	r.With(mw.Auth).Get("/", h.GetTeachers)

	router.Mount("/teachers", r)
}
