package sqs

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/require"
)

func Test_getQueueURL(t *testing.T) {
	testCases := map[string]struct {
		queueName string
		err       error
	}{
		"successful get queue url": {
			queueName: SqsQueueName,
			err:       nil,
		},
		"failed get queue url": {
			queueName: SqsErrQueueName,
			err:       errors.New(errMessageFailedGetUrl),
		},
	}

	for _, tc := range testCases {
		output, err := getQueueURL(&SqsMock{}, tc.queueName)

		if tc.err == nil {
			require.NoError(t, err)
			require.Equal(t, output.QueueUrl, aws.String(SqsQueueUrlPrefix+tc.queueName))
		} else {
			require.Error(t, err)
			require.Equal(t, tc.err.Error(), err.Error())
		}
	}
}

func TestDeleteSQSMessage(t *testing.T) {
	svc := &SQSService{
		Session:   &session.Session{},
		SQSClient: &SqsMock{},
	}

	testCases := map[string]struct {
		messageId string
		err       error
	}{
		"successful delete": {
			messageId: "1",
			err:       nil,
		},
		"failed delete": {
			messageId: ErrMessageId,
			err:       errors.New(ErrMessageFailedDelete),
		},
	}

	for _, tc := range testCases {
		err := svc.DeleteSQSMessage(tc.messageId)

		if tc.err == nil {
			require.NoError(t, err)
		} else {
			require.Equal(t, tc.err, err)
		}
	}
}

func TestGetSQSMessage(t *testing.T) {
	svc := &SQSService{
		Session:   &session.Session{},
		SQSClient: &SqsMock{},
	}

	testCases := map[string]struct {
		queueUrl string
		err      error
	}{
		"successful send": {
			queueUrl: SqsQueueUrlPrefix + SqsQueueName,
			err:      nil,
		},
		"failed send": {
			queueUrl: SqsQueueUrlPrefix + SqsErrQueueName,
			err:      errors.New(ErrMessageFailedReceive),
		},
	}

	for _, tc := range testCases {
		svc.QueueURL = &tc.queueUrl
		out, err := svc.GetSQSMessage(&SQSReceiveMsgConfig{})

		if tc.err == nil {
			require.NoError(t, err)
			require.NotNil(t, out)
			require.Equal(t, len(out.Messages), 1)
			require.Equal(t, out.Messages[0].Body, SqsMessageBody)
		} else {
			require.Equal(t, tc.err, err)
		}
	}
}
