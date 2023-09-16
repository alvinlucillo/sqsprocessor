package main

import (
	"os"

	"github.com/alvinlucillo/sqs-processor/internal/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

// main is the entrypoint to run the client app
func main() {
	// initializes logger to log to standard output
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Client starting")

	// initializes struct with values from env vars
	// errors out if required vars are missing
	var env client.Environment
	err := envconfig.Process("", &env)
	if err != nil {
		logger.Error().Err(err).Msg("Error initializing env")
		return
	}

	// logger.Info().Msgf("Env %v", env)

	sqsClient, err := client.NewClient(logger, env)
	if err != nil {
		logger.Error().Err(err).Msg("Error initializing client")
		return
	}

	if err := sqsClient.Run(); err != nil {
		logger.Error().Err(err).Msg("Error running client")
		return
	}
}
