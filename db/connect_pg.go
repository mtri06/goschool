package db

import (
	"fmt"
	"goschool/internal/env"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type DBConfig struct {
	Host        string
	Port        int
	User        string
	Password    string
	Name        string
	SSLMode     string
	ConnTimeout time.Duration
}

func ConfigFromEnv() DBConfig {
	return DBConfig{
		Host:        env.Env.PgHost,
		Port:        env.Env.PgPort,
		User:        env.Env.PgUser,
		Password:    env.Env.PgPassword,
		Name:        env.Env.PgDBName,
		SSLMode:     env.Env.PgSSLMode,
		ConnTimeout: env.Env.PgConnTimeout,
	}
}

func ConnectPostgres(cfg DBConfig) *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode, int(cfg.ConnTimeout.Seconds()),
	)
	pgCfg, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse postgres config")
	}

	pgCfg.RuntimeParams["TimeZone"] = "UTC"

	db := sqlx.NewDb(stdlib.OpenDB(*pgCfg), "pgx")

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	return db
}
