package db

import (
	"database/sql"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

func ConnectPostgres(url string) *sql.DB {
	cfg, err := pgx.ParseConfig(url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to parse postgres URL")
	}

	cfg.RuntimeParams["TimeZone"] = "UTC"

	db := stdlib.OpenDB(*cfg)

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	return db
}
