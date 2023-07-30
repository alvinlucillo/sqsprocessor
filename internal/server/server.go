package server

import (
	"context"
	"fmt"
	"net"

	"github.com/alvinlucillo/sqs-processor/internal/sqs"
	"github.com/alvinlucillo/sqs-processor/internal/types"
	pb "github.com/alvinlucillo/sqs-processor/protogen/sqs"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	packageName = "server"
)

type Server interface {
	ReceiveMessage(ctx context.Context, in *pb.SQSReceiveMessageRequest) (*pb.SQSReceiveMessageResponse, error)
	GracefulStop()
	Serve() error
}

type SQSServer struct {
	pb.SQSServiceServer
	SQSService *sqs.SQSService
	Logger     zerolog.Logger
	GrpcServer *grpc.Server
	Listener   net.Listener
}

func NewServer(logger zerolog.Logger, env types.Env) (Server, error) {
	l := logger.With().Str("package", packageName).Logger()

	sqsServer := &SQSServer{}

	sqsConfig := &sqs.SQSConfig{
		QueueName: env.QueueName,
		Profile:   env.Profile,
		Region:    env.Region,
		Logger:    logger,
	}

	sqsService, err := sqs.NewSQSService(sqsConfig)
	if err != nil {
		l.Err(err).Msg("Failed to initialize new SQS service")
		return nil, err
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", env.Port))
	if err != nil {
		l.Err(err).Msg("Failed to create listener")
		return nil, err
	}

	sqsServer.Logger = logger
	sqsServer.SQSService = sqsService
	sqsServer.GrpcServer = grpc.NewServer()
	sqsServer.Listener = listener

	pb.RegisterSQSServiceServer(sqsServer.GrpcServer, sqsServer)

	return sqsServer, nil
}

func (s *SQSServer) Serve() error {

	return s.GrpcServer.Serve(s.Listener)
}

func (s *SQSServer) GracefulStop() {
	l := s.Logger.With().Str("function", "GracefulStop").Logger()
	l.Info().Msg("Gracefully shutting down")

	s.GrpcServer.GracefulStop()
}

func (s *SQSServer) DeleteMessage(ctx context.Context, in *pb.SQSDeleteMessageRequest) (*emptypb.Empty, error) {
	l := s.Logger.With().Str("function", "DeleteMessage").Logger()

	l.Debug().Str("input", in.MessageID)

	return &emptypb.Empty{}, s.SQSService.DeleteSQSMessage(in.MessageID)
}

func (s *SQSServer) ReceiveMessage(ctx context.Context, in *pb.SQSReceiveMessageRequest) (*pb.SQSReceiveMessageResponse, error) {
	l := s.Logger.With().Str("function", "ReceiveMessage").Logger()

	l.Debug().Msgf("Received input: %v", in)

	sqsConfig := &sqs.SQSReceiveMsgConfig{
		VisibilityTimeout: in.VisibilityTimeout,
		WaitingTime:       in.WaitTime,
		MaximumMessages:   5,
	}

	messages, err := s.SQSService.GetSQSMessage(sqsConfig)
	if err != nil {
		l.Err(err).Msg("Failed to get SQS message")
		return nil, err
	}

	sqsReceiveResponse := make([]*pb.SQSResponseMessage, 0)

	for _, message := range messages.Messages {
		sqsReceiveResponse = append(sqsReceiveResponse, &pb.SQSResponseMessage{
			MessageID:   message.ID,
			MessageBody: message.Body,
		})
	}

	l.Debug().Msgf("Returned output: %v", sqsReceiveResponse)

	return &pb.SQSReceiveMessageResponse{
		Messages: sqsReceiveResponse,
	}, nil
}
