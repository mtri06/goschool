package api

import (
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/internal/api/routes"
	"goschool/pkg/httpx"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
)

type API struct {
	router *chi.Mux
}

func New(
	errMap httpx.APIErrorMap,
	authSvc handler.AuthService,
	studentSvc handler.StudentSvc,
	subjectSvc handler.SubjectSvc,
	teacherSvc handler.TeacherSvc,
) *API {
	// Init handlers
	teacherHandler := handler.NewTeacherHandler(teacherSvc, errMap)
	studentHandler := handler.NewStudentHandler(studentSvc, errMap)
	authHandler := handler.NewAuthHandler(authSvc, errMap)
	subjectHandler := handler.NewSubjectHandler(subjectSvc, errMap)

	// New router
	r := chi.NewRouter()

	// Init middlewares
	r.Use(mw.CORS())
	r.Use(chiMw.Timeout(20 * time.Second))
	r.Use(chiMw.Recoverer)
	r.Use(mw.Logging)
	r.NotFound(mw.NotFound())

	// Mount routes
	routes.MountTeacherRoutes(r, teacherHandler)
	routes.MountStudentRoutes(r, studentHandler)
	routes.MountSubjectRoutes(r, subjectHandler)
	routes.MountAuthRoutes(r, authHandler)

	return &API{router: r}
}

func (s *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}
