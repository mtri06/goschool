package main

import (
	"goschool/internal/db"
	"goschool/internal/env"
	"goschool/internal/server"
	"goschool/pkg/logger"

	"github.com/rs/zerolog/log"
)

func main() {
	env.Init()
	logger.Init()

	// Connect to Postgres
	dbClient := db.ConnectPostgres(db.DBConfig{
		Host:        env.Env.PgHost,
		Port:        env.Env.PgPort,
		User:        env.Env.PgUser,
		Password:    env.Env.PgPassword,
		Name:        env.Env.PgDBName,
		SSLMode:     env.Env.PgSSLMode,
		ConnTimeout: env.Env.PgConnTimeout,
	})
	log.Info().Msg("Connect to Postgres successfully")
	// Migrate database
	db.Migrate(dbClient.DB)

	server := server.New(dbClient)
	if err := server.Run(); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
	log.Info().Msg("Server stopped")
}
