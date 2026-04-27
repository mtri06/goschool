package api

import (
	"goschool/db"
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/internal/api/routes"
	"goschool/internal/env"
	"goschool/internal/repository"
	"goschool/internal/service"
	"goschool/pkg/logger"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func InitServer() http.Handler {
	env.Init()
	logger.Init()
	return NewServer(db.DBConfig{
		Host:        env.Env.PgHost,
		Port:        env.Env.PgPort,
		User:        env.Env.PgUser,
		Password:    env.Env.PgPassword,
		Name:        env.Env.PgDBName,
		SSLMode:     env.Env.PgSSLMode,
		ConnTimeout: env.Env.PgConnTimeout,
	})
}

// NewServer wires the full application given a database URL.
// env.Env must be populated before calling this (InitServer handles that for production;
// integration tests set env.Env fields directly).
func NewServer(dbCfg db.DBConfig) http.Handler {
	// Connect to Postgres
	dbClient := db.ConnectPostgres(dbCfg)
	log.Info().Msg("Connect to Postgres successfully")

	// Migrate database
	db.Migrate(dbClient.DB)

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
	teacherSvc := service.NewTeacherService(userRepo, teacherRepo, subjectRepo, userSvc)
	studentSvc := service.NewStudentService(userRepo, studentRepo, classRepo, userSvc)
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
