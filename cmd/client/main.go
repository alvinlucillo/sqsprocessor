package main

import (
	"os"

	"github.com/alvinlucillo/sqs-processor/internal/client"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

func main() {

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Client starting")

	var env client.Environment
	err := envconfig.Process("myapp", &env)
	if err != nil {
		logger.Error().Err(err).Msg("Error initializing env")
		return
	}

	logger.Info().Msgf("Env %v", env)

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
