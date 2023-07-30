package main

import (
	"os"
	"os/signal"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/alvinlucillo/sqs-processor/internal/server"
	"github.com/alvinlucillo/sqs-processor/internal/types"

	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Server starting")

	env, err := getEnvironmentValues()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to retrieve environment variables")
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

func getEnvironmentValues() (types.Env, error) {
	env := types.Env{}

	t := reflect.TypeOf(env)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("def")
		// first element (0) - env name
		// second elemtn (1) - default value if env not set
		splitTag := strings.Split(tag, ",")

		envValue, found := os.LookupEnv(splitTag[0])
		if found {
			reflect.ValueOf(&env).Elem().FieldByName(field.Name).Set(reflect.ValueOf(envValue))
		} else {
			if field.Type == reflect.TypeOf(int(0)) {
				intVal, err := strconv.Atoi(splitTag[1])
				if err != nil {
					return env, err
				}
				reflect.ValueOf(&env).Elem().FieldByName(field.Name).Set(reflect.ValueOf(intVal))
			} else {
				reflect.ValueOf(&env).Elem().FieldByName(field.Name).Set(reflect.ValueOf(splitTag[1]))
			}

		}
	}

	return env, nil
}
