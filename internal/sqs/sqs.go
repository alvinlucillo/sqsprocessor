package sqs

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/rs/zerolog"
)

const (
	packageName = "sqs"
)

type SQSService struct {
	Session   *session.Session
	SQSClient *sqs.SQS
	QueueURL  *string
	Logger    zerolog.Logger
}

// NewSQSService - creates new SQSService
func NewSQSService(config *SQSConfig) (*SQSService, error) {
	l := config.Logger.With().Str("package", packageName).Logger()

	sqsService := &SQSService{}
	sqsService.Logger = l

	l = l.With().Str("function", "NewSQSService").Logger()

	session, err := session.NewSession(&aws.Config{
		Region: aws.String(config.Region),
	})

	if err != nil {
		l.Err(err).Msg("Failed to initialize new session")

		return nil, err
	}

	sqsClient := sqs.New(session)

	queueURL, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &config.QueueName,
	})
	if err != nil {
		l.Err(err).Msg("Failed to get queue URL")
		return nil, err
	}

	sqsService.Session = session
	sqsService.QueueURL = queueURL.QueueUrl
	sqsService.SQSClient = sqsClient

	return sqsService, nil
}

func (s *SQSService) DeleteSQSMessage(id string) error {
	input := &sqs.DeleteMessageInput{
		QueueUrl:      s.QueueURL,
		ReceiptHandle: aws.String(id),
	}

	_, err := s.SQSClient.DeleteMessage(input)

	return err
}

func (s *SQSService) GetSQSMessage(sqsConfig *SQSReceiveMsgConfig) (*SQSResult, error) {
	l := s.Logger.With().Str("function", "GetSQSMessage").Logger()

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            s.QueueURL,
		MaxNumberOfMessages: aws.Int64(sqsConfig.MaximumMessages),
		VisibilityTimeout:   aws.Int64(sqsConfig.VisibilityTimeout),
		WaitTimeSeconds:     aws.Int64(sqsConfig.WaitingTime),
	}

	result, err := s.pollForMsgs(s.Session, input)
	if err != nil {
		l.Err(err).Msgf("Failed to poll for messages")
		return nil, err
	}

	messages := make([]SQSResultMessage, 0)

	if len(result) > 0 {
		for _, msg := range result {
			messages = append(messages, SQSResultMessage{ID: *msg.ReceiptHandle, Body: *msg.Body})
			fmt.Println(*msg.ReceiptHandle)
		}
	}

	return &SQSResult{Messages: messages}, nil
}

func (s *SQSService) pollForMsgs(sess *session.Session, sqsMessageInput *sqs.ReceiveMessageInput) ([]*sqs.Message, error) {
	l := s.Logger.With().Str("function", "pollForMsgs").Logger()

	msgResult, err := s.SQSClient.ReceiveMessage(sqsMessageInput)

	if err != nil {
		l.Err(err).Msgf("Failed to query messages from SQS")
		return nil, err
	}

	return msgResult.Messages, nil
}
