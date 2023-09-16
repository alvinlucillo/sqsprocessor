package sqs

// package is used for unit test
// mocks aws sdk's sqs functions for testing purposes

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
)

// default values
const (
	SqsQueueName         = "queue-1"
	SqsErrQueueName      = "err-queue"
	SqsQueueUrlPrefix    = "https://sqs.us-east-1.amazonaws.com/12345/"
	SqsMessageRcptHandle = "message-1"
	SqsMessageId         = "message-id-1"
	SqsMessageBody       = "message-body"

	ErrMessageId            = "error-id"
	ErrMessageFailedDelete  = "failed deleting message"
	errMessageFailedGetUrl  = "failed getting url"
	ErrMessageFailedReceive = "failed receiving message"
)

type SqsMock struct {
	sqsiface.SQSAPI
	deleteMessageOutput *sqs.DeleteMessageOutput
}

// DeleteMessage -- mocks sqs DeleteMessage
func (s SqsMock) DeleteMessage(in *sqs.DeleteMessageInput) (*sqs.DeleteMessageOutput, error) {

	if *in.ReceiptHandle == ErrMessageId {
		return s.deleteMessageOutput, errors.New(ErrMessageFailedDelete)
	}

	return s.deleteMessageOutput, nil
}

// GetQueueUrl -- mocks sqs GetQueueUrl
func (s SqsMock) GetQueueUrl(in *sqs.GetQueueUrlInput) (*sqs.GetQueueUrlOutput, error) {

	if *in.QueueName == SqsErrQueueName {
		return nil, errors.New(errMessageFailedGetUrl)
	}

	url := SqsQueueUrlPrefix + *in.QueueName
	return &sqs.GetQueueUrlOutput{
		QueueUrl: &url,
	}, nil
}

// ReceiveMessage -- mocks sqs ReceiveMessage
func (s SqsMock) ReceiveMessage(in *sqs.ReceiveMessageInput) (*sqs.ReceiveMessageOutput, error) {
	if *in.QueueUrl == SqsQueueUrlPrefix+SqsErrQueueName {
		return nil, errors.New(ErrMessageFailedReceive)
	}

	out := &sqs.ReceiveMessageOutput{
		Messages: []*sqs.Message{{ReceiptHandle: aws.String(SqsMessageRcptHandle), Body: aws.String(SqsMessageBody)}},
	}

	return out, nil
}
