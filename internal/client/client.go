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
	Conn              *grpc.ClientConn
	Client            pb.SQSServiceClient
	Logger            zerolog.Logger
	PollingInterval   int
	VisibilityTimeout int
	WaitTime          int
	ErrorRateLimit    int
	MaximumMessages   int
}

type Environment struct {
	// sqsservice port the client will connect to
	SQSServicePort int `required:"true" default:"50051"`
	// the limit during which the polling process will terminate
	ErrorRateLimit int `required:"true" default:"10"`
	// number of seconds between each client's poll action
	PollingInterval int `required:"true" default:"5"`
	// number of seconds the received message becomes invisible to other consumers
	// passed to the sqsservice
	VisibilityTimeout int `required:"true" default:"5"`
	// number of seconds the sqs api will wait before retrieving messages
	// passed to the sqsservice
	// 0 means short polling (more costly, quicker retrieval interval)
	// higher values means long polling (less costly, longer retrieval interval)
	WaitTime int `required:"true" default:"5"`
	// number of messages the sqs api will retrieve from the queue
	// passed to the sqsservice
	MaximumMessages int `required:"true" default:"5"`
}

// NewClient - initializes a new client app
func NewClient(logger zerolog.Logger, env Environment) (Client, error) {
	logger = logger.With().Str("package", packageName).Logger()
	l := logger.With().Str("function", "NewClient").Logger()

	sqsClient := &SQSClient{Logger: logger, PollingInterval: env.PollingInterval,
		VisibilityTimeout: env.VisibilityTimeout, WaitTime: env.WaitTime, ErrorRateLimit: env.ErrorRateLimit, MaximumMessages: env.MaximumMessages}

	// establishing connection to sqsservice
	target := fmt.Sprintf("localhost:%v", env.SQSServicePort)
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

// Run - drives the polling process
// only stops when error or system terminates it
func (s *SQSClient) Run() error {
	l := s.Logger.With().Str("function", "Run").Logger()

	ctx := context.Background()

	// establishes parameters for the sqsservice
	req := &pb.SQSReceiveMessageRequest{VisibilityTimeout: int64(s.VisibilityTimeout), WaitTime: int64(s.WaitTime), MaximumNumberOfMessages: int64(s.MaximumMessages)}
	pollCounter := 0
	errCounter := 0
	processed := 0

	// logs summary and closes connection before leaving the function
	defer func() {
		l.Debug().Msgf("Number of processed messages %v", processed)
		l.Debug().Msgf("Number of polls made %v", pollCounter)
		l.Debug().Msgf("Number of errors encountered %v", errCounter)

		if err := s.Conn.Close(); err != nil {
			l.Error().Err(err).Msg("Unable to close connection")
		}
	}()

	// pools for sqs messages indefinitely until error rate exceeds limit
	// once received, each message is deleted from the queue
	for {
		pollCounter++

		l.Info().Msgf("Polling count: %v", pollCounter)

		resp, err := s.Client.ReceiveMessage(ctx, req)
		if err != nil {
			l.Error().Err(err).Msg("Unable to receive message from sqs")
			errCounter++

			if errCounter > s.ErrorRateLimit {
				return fmt.Errorf("number of errors exceeded limit: %v", errCounter)
			}

			continue
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

					if errCounter > s.ErrorRateLimit {
						return fmt.Errorf("number of errors exceeded limit: %v", errCounter)
					}
				}

				l.Info().Msgf("Message deleted successfully: %v", msg.MessageID)
			}
		} else {
			l.Info().Msg("No messages received")
		}

		// waits for the set time before going through another polling
		time.Sleep(time.Duration(s.PollingInterval) * time.Second)
	}
}
