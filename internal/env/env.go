package env

import (
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

type envConfig struct {
	Environment           string
	DBURL                 string
	AllowedOrigins        []string
	JWTSecret             string
	JWTAccessExpiresMins  int
	JWTRefreshExpiresDays int
	AdminUsername         string
	AdminPassword         string
}

var Env envConfig

func Init() {
	if err := godotenv.Load(); err != nil {
		log.Fatal().Msgf("Error loading .env file: %v", err)
	}

	Env = envConfig{
		Environment:           envOrDefault("ENVIRONMENT", "dev"),
		DBURL:                 envOrPanic("DATABASE_URL"),
		AllowedOrigins:        strings.Split(envOrPanic("CORS_ALLOWED_ORIGINS"), ","),
		JWTSecret:             envOrPanic("JWT_SECRET"),
		JWTAccessExpiresMins:  envIntOrDefault("JWT_ACCESS_EXPIRES_MINS", 15),
		JWTRefreshExpiresDays: envIntOrDefault("JWT_REFRESH_EXPIRES_DAYS", 7),
		AdminUsername:         envOrPanic("ADMIN_USERNAME"),
		AdminPassword:         envOrPanic("ADMIN_PASSWORD"),
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
