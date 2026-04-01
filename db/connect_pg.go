package db

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func ConnectPostgres(url string) *sqlx.DB {
	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse postgres URL")
	}

	cfg.RuntimeParams["TimeZone"] = "UTC"

	db := sqlx.NewDb(stdlib.OpenDB(*cfg), "pgx")

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	return db
}
