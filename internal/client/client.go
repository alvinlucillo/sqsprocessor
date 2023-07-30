package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/alvinlucillo/sqs-processor/protogen/sqs"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	packageName = "client"
)

type Client interface {
}

type SQSClient struct {
	Conn            *grpc.ClientConn
	Client          pb.SQSServiceClient
	Logger          zerolog.Logger
	PollingInterval int
}

type Environment struct {
	Port            int `required:"true" default:"50051"`
	PollingInterval int `required:"true" default:"5"`
}

func NewClient(logger zerolog.Logger, env Environment) (Client, error) {
	l := logger.With().Str("package", packageName).Logger()

	l = l.With().Str("function", "NewClient").Logger()

	sqsClient := &SQSClient{Logger: l}

	conn, err := grpc.Dial(fmt.Sprint(env.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error().Err(err).Msg("Failed to connect")
	}

	defer conn.Close()

	client := pb.NewSQSServiceClient(conn)

	sqsClient.Client = client
	sqsClient.Conn = conn

	return sqsClient, nil
}

func (s *SQSClient) Run() error {

	l := s.Logger.With().Str("function", "Run").Logger()

	ctx := context.Background()

	req := &pb.SQSReceiveMessageRequest{VisibilityTimeout: 5, WaitTime: 5}

	pollCounter := 0

	for {
		pollCounter++

		l.Info().Msgf("Polling %v time(s)", pollCounter)

		resp, err := s.Client.ReceiveMessage(ctx, req)
		if err != nil {
			l.Error().Err(err).Msg("Unable to receive message from sqs")
		}

		if resp != nil {
			l.Info().Msgf("Received %v message(s)", len(resp.Messages))

			for _, msg := range resp.Messages {
				deleteReq := &pb.SQSDeleteMessageRequest{MessageID: msg.MessageID}
				_, err := s.Client.DeleteMessage(ctx, deleteReq)

				if err != nil {
					l.Error().Err(err).Msg("Unable to delete ")
					continue
				}

				l.Info().Msgf("Message id %v deleted successfully", msg.MessageID)
			}
		}

		time.Sleep(time.Duration(s.PollingInterval) * time.Second)
	}
}
