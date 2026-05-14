package api

import (
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/internal/api/routes"
	"goschool/internal/repository"
	"goschool/internal/service"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/jmoiron/sqlx"
)

// NewServer wires up repositories, services, handlers, and routes, and returns the http.Handler for the server.
func NewServer(dbClient *sqlx.DB) http.Handler {

	r := chi.NewRouter()

	// Init middlewares
	r.Use(mw.CORS())
	r.Use(chiMw.Timeout(20 * time.Second))
	r.Use(chiMw.Recoverer)
	r.Use(mw.Logging)
	r.NotFound(mw.NotFound())

	// Init repositories
	userRepo := repository.NewUserRepository(dbClient)
	teacherRepo := repository.NewTeacherRepository(dbClient)
	studentRepo := repository.NewStudentRepository(dbClient)
	subjectRepo := repository.NewSubjectRepository(dbClient)
	classRepo := repository.NewClassRepository(dbClient)
	tokenRepo := repository.NewTokenRepository(dbClient)

	// Init services
	userSvc := service.NewUserService(userRepo)
	teacherSvc := service.NewTeacherService(userRepo, teacherRepo, subjectRepo)
	studentSvc := service.NewStudentService(userRepo, studentRepo, classRepo)
	authSvc := service.NewAuthService(userRepo, tokenRepo)

	// Seed admin user
	userSvc.SeedAdminUser()

	// Init handlers
	errMap := handler.NewErrorMap()
	teacherHandler := handler.NewTeacherHandler(teacherSvc, errMap)
	studentHandler := handler.NewStudentHandler(studentSvc, errMap)
	authHandler := handler.NewAuthHandler(authSvc, errMap)

	// Mount routes
	routes.MountTeacherRoutes(r, teacherHandler)
	routes.MountStudentRoutes(r, studentHandler)
	routes.MountAuthRoutes(r, authHandler)

	return r
}
