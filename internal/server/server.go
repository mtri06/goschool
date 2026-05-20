package server

import (
	"goschool/internal/api"
	"goschool/internal/repository"
	"goschool/internal/service"
	"goschool/pkg/httpx"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type Server struct {
	api *api.API
}

func New(dbClient *sqlx.DB) *Server {

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
	subjectSvc := service.NewSubjectService(subjectRepo)

	// Seed admin user if not exists
	userSvc.SeedAdminUser()

	api := api.New(
		errMap,
		authSvc,
		studentSvc,
		subjectSvc,
		teacherSvc,
	)

	return &Server{api: api}
}

func (s *Server) Run() error {
	log.Info().Msg("Starting server on :8080")
	return http.ListenAndServe(":8080", s.api)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.api.ServeHTTP(w, r)
}

var errMap = map[error]*httpx.APIError{
	service.ErrValidationFailed: httpx.ErrBadRequest,
	service.ErrNotFound:         httpx.ErrNotFound,
	service.ErrUnauthorized:     httpx.ErrUnauthorized,
	service.ErrForbidden:        httpx.ErrForbidden,
}
