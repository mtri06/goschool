package routes

import (
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/pkg/constant"

	"github.com/go-chi/chi/v5"
)

func MountSubjectRoutes(router chi.Router, h *handler.SubjectHandler) {
	r := chi.NewRouter()

	r.With(mw.Auth, mw.RequireRole(constant.RoleAdmin)).Post("/", h.CreateSubject)

	router.Mount("/subjects", r)
}
