package sqsservice

import (
	"context"
	"errors"
	"testing"

	pb "github.com/alvinlucillo/sqs-processor/protogen/sqs"

	"github.com/alvinlucillo/sqs-processor/internal/sqs"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/require"
)

func TestDeleteMessage(t *testing.T) {

	svc := &sqs.SQSService{
		Session:   &session.Session{},
		SQSClient: &sqs.SqsMock{},
	}

	server := &SQSServer{
		SQSService: svc,
	}

	testCases := map[string]struct {
		messageId string
		err       error
	}{
		"successful delete": {
			messageId: sqs.SqsMessageId,
			err:       nil,
		},
		"failed delete": {
			messageId: sqs.ErrMessageId,
			err:       errors.New(sqs.ErrMessageFailedDelete),
		},
	}

	for _, tc := range testCases {
		_, err := server.DeleteMessage(context.Background(), &pb.SQSDeleteMessageRequest{MessageID: tc.messageId})

		if tc.err == nil {
			require.NoError(t, err)
		} else {
			require.Equal(t, tc.err, err)
		}
	}
}

func TestReceiveMessage(t *testing.T) {

	svc := &sqs.SQSService{
		Session:   &session.Session{},
		SQSClient: &sqs.SqsMock{},
	}

	server := &SQSServer{
		SQSService: svc,
	}

	testCases := map[string]struct {
		queueName string
		err       error
	}{
		"successful receive": {
			queueName: sqs.SqsQueueUrlPrefix + sqs.SqsQueueName,
			err:       nil,
		},
		"failed receive": {
			queueName: sqs.SqsQueueUrlPrefix + sqs.SqsErrQueueName,
			err:       errors.New(sqs.ErrMessageFailedReceive),
		},
	}

	for _, tc := range testCases {
		svc.QueueURL = aws.String(tc.queueName)
		_, err := server.ReceiveMessage(context.Background(), &pb.SQSReceiveMessageRequest{})

		if tc.err == nil {
			require.NoError(t, err)
		} else {
			require.Equal(t, tc.err, err)
		}
	}
}
