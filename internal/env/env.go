package env

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type envConfig struct {
	Environment string

	PgHost        string
	PgPort        string
	PgUser        string
	PgPassword    string
	PgDBName      string
	PgSSLMode     string
	PgConnTimeout time.Duration

	AllowedOrigins        []string
	JWTSecret             string
	JWTAccessExpiresMins  int
	JWTRefreshExpiresDays int
	AdminUsername         string
	AdminPassword         string
}

var Env envConfig

func Init(envFiles ...string) {
	if err := godotenv.Load(envFiles...); err != nil {
		log.Warn().Msg("No .env file found, relying on environment variables")
	}

	Env = envConfig{
		Environment: envOrDefault("ENVIRONMENT", "dev"),

		PgHost:        envOrPanic("PG_HOST"),
		PgPort:        envOrPanic("PG_PORT"),
		PgUser:        envOrPanic("PG_USER"),
		PgPassword:    envOrPanic("PG_PASSWORD"),
		PgDBName:      envOrPanic("PG_DB_NAME"),
		PgSSLMode:     envOrPanic("PG_SSL_MODE"),
		PgConnTimeout: envDurationOrDefault("PG_CONN_TIMEOUT", 10*time.Second),

		AllowedOrigins: strings.Split(envOrPanic("CORS_ALLOWED_ORIGINS"), ","),

		JWTSecret:             envOrPanic("JWT_SECRET"),
		JWTAccessExpiresMins:  envIntOrDefault("JWT_ACCESS_EXPIRES_MINS", 15),
		JWTRefreshExpiresDays: envIntOrDefault("JWT_REFRESH_EXPIRES_DAYS", 7),

		AdminUsername: envOrPanic("ADMIN_USERNAME"),
		AdminPassword: envOrPanic("ADMIN_PASSWORD"),
	}
}

func envOrPanic(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Panic().Msgf("%v env is not set", key)
	}
	return val
}

func envOrDefault(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func envIntOrDefault(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		log.Panic().Msgf("%v env is not a valid integer: %v", key, err)
	}
	return intVal
}

func envDurationOrDefault(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	duration, err := time.ParseDuration(val)
	if err != nil {
		log.Panic().Msgf("%v env is not a valid duration: %v", key, err)
	}
	return duration
}
