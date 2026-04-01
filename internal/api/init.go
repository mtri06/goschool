package api

import (
	"goschool/db"
	"goschool/internal/api/handler"
	mw "goschool/internal/api/middleware"
	"goschool/internal/api/routes"
	"goschool/internal/env"
	"goschool/internal/repository"
	"goschool/internal/services"
	"goschool/pkg/httpx"
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

	// Connect to Postgres
	dbClient := db.ConnectPostgres(env.Env.DBURL)
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
	subjectRepo := repository.NewSubjectRepository(dbClient)
	tokenRepo := repository.NewTokenRepository(dbClient)

	// Init services
	userSvc := services.NewUserService(userRepo)
	teacherSvc := services.NewTeacherService(userRepo, subjectRepo, userSvc)
	authSvc := services.NewAuthService(userRepo, tokenRepo)

	// Seed admin user
	userSvc.SeedAdminUser()

	// Init handlers
	teacherHandler := handler.NewTeacherHandler(teacherSvc, serviceErrMapping)
	authHandler := handler.NewAuthHandler(authSvc, serviceErrMapping)

	// Mount routes
	routes.MountTeacherRoutes(r, teacherHandler)
	routes.MountAuthRoutes(r, authHandler)

	return r
}

var serviceErrMapping = map[error]httpx.APIError{
	httpx.ErrInvalidBody:           httpx.ErrBadRequest,
	httpx.ErrInvalidQuery:          httpx.ErrBadRequest,
	services.ErrValidationFailed:   httpx.ErrBadRequest,
	services.ErrNotFound:           httpx.ErrNotFound,
	services.ErrInvalidCredentials: httpx.ErrUnauthorized,
	services.ErrUnauthorized:       httpx.ErrUnauthorized,
}
