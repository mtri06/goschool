package db

import (
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog/log"
)

func ConnectPostgres(url string) *sql.DB {
	db, err := sql.Open("pgx", url)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to postgres")
	}

	if err := db.Ping(); err != nil {
		log.Fatal().Err(err).Msg("failed to ping postgres")
	}

	return db
}
