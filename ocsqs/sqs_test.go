package ocsqs

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/ocawstest"
	"go.opencensus.io/trace"
)

func TestMain(m *testing.M) {
	trace.ApplyConfig(trace.Config{
		IDGenerator: ocawstest.NewTestIDGenerator(),
	})

	os.Exit(m.Run())
}

func TestDefaultFormatSpanName(t *testing.T) {
	type TestCase struct {
		tName   string
		message *sqs.Message
		name    string
	}
	tt := []TestCase{
		{
			tName:   "empty message",
			message: &sqs.Message{},
			name:    "sqs.Message/unknwonMessageId",
		},
		{
			tName: "with message id",
			message: &sqs.Message{
				MessageId: aws.String("some-message-id"),
			},
			name: "sqs.Message/some-message-id",
		},
		{
			tName: "with topic name",
			message: &sqs.Message{
				MessageId: aws.String("some-message-id"),
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					ocaws.TraceTopicName: &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String("Foo"),
					},
				},
			},
			name: "sqs.Message/Foo/some-message-id",
		},
		{
			tName: "with queue url",
			message: &sqs.Message{
				MessageId: aws.String("some-message-id"),
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					ocaws.TraceQueueURL: &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String("https://sqs.eu-west-1.amazonaws.com/123456789101112/Bar"),
					},
				},
			},
			name: "sqs.Message/123456789101112/Bar/some-message-id",
		},
		{
			tName: "with queue url and topic",
			message: &sqs.Message{
				MessageId: aws.String("some-message-id"),
				MessageAttributes: map[string]*sqs.MessageAttributeValue{
					ocaws.TraceQueueURL: &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String("https://sqs.eu-west-1.amazonaws.com/123456789101112/Bar"),
					},
					ocaws.TraceTopicName: &sqs.MessageAttributeValue{
						DataType:    aws.String("String"),
						StringValue: aws.String("Foo"),
					},
				},
			},
			name: "sqs.Message/Foo/123456789101112/Bar/some-message-id",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			assert.Equal(t, tc.name, DefaultFormatSpanName(tc.message))

			t.Parallel()
		})
	}
}
