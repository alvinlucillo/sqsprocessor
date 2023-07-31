package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/alvinlucillo/sqs-processor/internal/server"
	"github.com/alvinlucillo/sqs-processor/internal/types"
	"github.com/kelseyhightower/envconfig"

	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Server starting")

	var env types.ServerEnvironment
	err := envconfig.Process("myapp", &env)
	if err != nil {
		logger.Error().Err(err).Msg("Error initializing env")
		return
	}

	logger.Debug().Msgf("Environment variables %v", env)

	s, err := server.NewServer(logger, env)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create server")
	}

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err = s.Serve(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start the server")
		}
	}()

	<-shutdownChan
	s.GracefulStop()

	logger.Info().Msg("Server stopped")
}
