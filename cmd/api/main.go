package main

import (
	"goschool/db"
	"goschool/internal/api"
	"goschool/internal/env"
	"goschool/pkg/logger"
	"net/http"

	"github.com/rs/zerolog/log"
)

func main() {
	env.Init()
	logger.Init()

	// Connect to Postgres
	dbClient := db.ConnectPostgres(db.ConfigFromEnv())
	log.Info().Msg("Connect to Postgres successfully")
	// Migrate database
	db.Migrate(dbClient.DB)

	server := api.NewServer(dbClient)

	log.Info().Msg("Server started at http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", server); err != nil {
		log.Fatal().Msgf("Could not start http server: %v", err)
	}
}
