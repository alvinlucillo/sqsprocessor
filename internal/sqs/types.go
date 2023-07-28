package sqs

import "github.com/rs/zerolog"

type SQSReceiveMsgConfig struct {
	VisibilityTimeout int64
	WaitingTime       int64
	MaximumMessages   int64
}

type SQSResultMessage struct {
	ID   string
	Body string
}

type SQSResult struct {
	Messages []SQSResultMessage
}

type SQSConfig struct {
	QueueName string
	Profile   string
	Logger    zerolog.Logger
	Region    string
}
