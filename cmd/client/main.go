package main

import (
	"os"

	"github.com/alvinlucillo/sqs-processor/internal/client"
	pb "github.com/alvinlucillo/sqs-processor/protogen/sqs"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
)

type Client struct {
	pb.SQSServiceClient
}

var addr string = "0.0.0.0:50051"

func main() {

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Caller().Msg("Client starting")

	var env client.Environment
	err := envconfig.Process("myapp", &env)
	if err != nil {
		logger.Error().Err(err).Msg("Error env")

		return
	}

	logger.Info().Msgf("Env %v", env)

	// var conn *grpc.ClientConn

	// conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// if err != nil {
	// 	log.Fatalf("Failed to connect: %v\n", err)
	// }

	// // perform close at the end of the function
	// defer conn.Close()

	// c := pb.NewSQSServiceClient(conn)

	// res, err := c.ReceiveMessage(context.Background(), &pb.SQSReceiveMessageRequest{VisibilityTimeout: 5, WaitTime: 5})
	// if err != nil {
	// 	log.Fatalf("Failed to call receive message: %v\n", err)
	// }

	// log.Printf("Message: %s\n", res.Messages)

	// res1, err := c.DeleteMessage(context.Background(), &pb.SQSDeleteMessageRequest{MessageID: res.Messages[0].MessageID})
	// if err != nil {
	// 	log.Fatalf("Failed to call delete message: %v\n", err)
	// }

	// log.Printf("Result: %s\n", res1)

}
