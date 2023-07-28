package main

import (
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/alvinlucillo/sqs-processor/internal/server"
	"github.com/alvinlucillo/sqs-processor/internal/types"

	"github.com/rs/zerolog"
)

// var addr string = "0.0.0.0:50051"

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Server started")

	// lis, err := net.Listen("tcp", addr)
	// if err != nil {
	// 	logger.Fatal().Msgf("Failed to listen on: %v\n", err)
	// }

	// logger.Info().Msgf("Listening on %s", addr)

	// opts := []grpc.ServerOption{}
	// s := grpc.NewServer(opts...)

	env, err := getEnvironmentValues()
	if err != nil {
		logger.Fatal().Msgf("Failed to retrieve environment variables: %v", err)
	}

	logger.Debug().Msgf("Environment variables %v", env)

	s, err := server.NewServer(logger, env)
	if err != nil {
		logger.Fatal().Msgf("Failed to create server: %v", err)
	}

	if err = s.Serve(); err != nil {
		logger.Fatal().Msgf("Failed to serve: %v", err)
	}

	// pb.RegisterSQSServiceServer(s, server)
	// if err = s.Serve(lis); err != nil {
	// 	logger.Fatal().Msgf("Failed to serve: %v", err)
	// }

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
