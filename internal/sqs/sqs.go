package sqs

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"github.com/rs/zerolog"
)

const (
	packageName = "sqs"
)

type SQSService struct {
	Session   *session.Session
	SQSClient sqsiface.SQSAPI
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
		Region:      aws.String(config.Region),
		Credentials: credentials.NewStaticCredentials("key", "secret", "token"),
	})

	if err != nil {
		l.Err(err).Msg("Failed to create new session")

		return nil, err
	}

	sqsClient := sqs.New(session)

	queueURL, err := getQueueURL(sqsClient, config.QueueName)
	if err != nil {
		l.Err(err).Msg("Failed to get queue URL")
		return nil, err
	}

	sqsService.Session = session
	sqsService.QueueURL = queueURL.QueueUrl
	sqsService.SQSClient = sqsClient

	return sqsService, nil
}

func getQueueURL(sqsClient sqsiface.SQSAPI, queueName string) (*sqs.GetQueueUrlOutput, error) {
	queueURL, err := sqsClient.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	return queueURL, err
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

	result, err := s.pollMessages(s.Session, input)
	if err != nil {
		l.Err(err).Msgf("Failed to poll for messages")
		return nil, err
	}

	messages := make([]SQSResultMessage, 0)

	if len(result) > 0 {
		for _, msg := range result {
			messages = append(messages, SQSResultMessage{ID: *msg.ReceiptHandle, Body: *msg.Body})
		}
	}

	return &SQSResult{Messages: messages}, nil
}

func (s *SQSService) pollMessages(sess *session.Session, sqsMessageInput *sqs.ReceiveMessageInput) ([]*sqs.Message, error) {
	l := s.Logger.With().Str("function", "pollMessages").Logger()

	msgResult, err := s.SQSClient.ReceiveMessage(sqsMessageInput)

	if err != nil {
		l.Err(err).Msgf("Failed to query messages from SQS")
		return nil, err
	}

	return msgResult.Messages, nil
}
