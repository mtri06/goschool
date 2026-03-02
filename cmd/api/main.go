package main

import (
	"goschool/internal/api"
	"net/http"

	"github.com/rs/zerolog/log"
)

func main() {
	server := api.InitServer()
	log.Info().Msg("Server started at http://localhost:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", server); err != nil {
		log.Fatal().Msgf("Could not start http server: %v", err)
	}
}
