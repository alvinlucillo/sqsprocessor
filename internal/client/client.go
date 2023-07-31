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
	Run() error
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
	logger = logger.With().Str("package", packageName).Logger()
	l := logger.With().Str("function", "NewClient").Logger()

	sqsClient := &SQSClient{Logger: logger, PollingInterval: env.PollingInterval}

	target := fmt.Sprintf("localhost:%v", env.Port)
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		l.Error().Err(err).Msg("Failed to connect")
		return nil, err
	}

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
	errCounter := 0
	processed := 0

	defer func() {
		l.Debug().Msgf("Number of processed messages %v", processed)
		l.Debug().Msgf("Number of polls made %v", pollCounter)
		l.Debug().Msgf("Number of errors encountered %v", errCounter)

		if err := s.Conn.Close(); err != nil {
			l.Error().Err(err).Msg("Unable to close connection")
		}
	}()

	for {
		pollCounter++

		l.Info().Msgf("Polling count: %v", pollCounter)

		resp, err := s.Client.ReceiveMessage(ctx, req)
		if err != nil {
			l.Error().Err(err).Msg("Unable to receive message from sqs")
			errCounter++

			if errCounter > 10 {
				goto exit
			}
		}

		if resp != nil {
			l.Info().Msgf("Received %v message(s)", len(resp.Messages))

			for _, msg := range resp.Messages {
				l.Info().Msgf("Deleting message %v", msg)

				deleteReq := &pb.SQSDeleteMessageRequest{MessageID: msg.MessageID}
				_, err := s.Client.DeleteMessage(ctx, deleteReq)

				if err != nil {
					l.Error().Err(err).Msgf("Unable to delete message: %v", msg.MessageID)
					errCounter++

					if errCounter > 10 {
						goto exit
					}
				}

				l.Info().Msgf("Message deleted successfully: %v", msg.MessageID)
			}
		} else {
			l.Info().Msg("No messages received")
		}

		time.Sleep(time.Duration(s.PollingInterval) * time.Second)
	}

exit:
	return fmt.Errorf("number of errors exceeded limit: %v", errCounter)
}
