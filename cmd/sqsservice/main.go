package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/alvinlucillo/sqs-processor/internal/sqsservice"
	"github.com/kelseyhightower/envconfig"

	"github.com/rs/zerolog"
)

// main is the entrypoint to run the sqsservice app
func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Server starting")

	// initializes struct with values from env vars
	// errors out if required vars are missing
	var env sqsservice.Environment
	err := envconfig.Process("app", &env)
	if err != nil {
		logger.Error().Err(err).Msg("Error initializing env")
		return
	}

	// logger.Debug().Msgf("Environment variables %v", env)

	s, err := sqsservice.NewServer(logger, env)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create server")
	}

	// channel to receive the signal if the program is terminated
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, syscall.SIGINT, syscall.SIGTERM)

	// runs the blocking function Server in a goroutine
	// so server can be gracefully stopped when signal is received
	go func() {
		if err = s.Serve(); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start the server")
		}
	}()

	// waits for signal before gracefully stopping the server
	<-shutdownChan
	s.GracefulStop()

	logger.Info().Msg("Server stopped")
}
