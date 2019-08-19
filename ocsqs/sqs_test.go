package ocsqs

import (
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/ocawstest"
	"go.krak3n.codes/ocaws/propagation/propagationtest"
	"go.opencensus.io/trace"
)

func TestMain(m *testing.M) {
	trace.ApplyConfig(trace.Config{
		IDGenerator: ocawstest.NewTestIDGenerator(),
	})

	os.Exit(m.Run())
}

func TestWithPropagator(t *testing.T) {
	p := &propagationtest.TestPropator{}
	c := New(nil, WithPropagator(p))

	assert.Equal(t, c.Propagator, p)
}

func TestWithStartOptions(t *testing.T) {
	s := trace.StartOptions{
		SpanKind: trace.SpanKindClient,
	}
	c := New(nil, WithStartOptions(s))

	assert.Equal(t, c.StartOptions, s)
}

func TestWithGetStartOptions(t *testing.T) {
	var fn GetStartOptionsFunc = GetStartOptionsFunc(func(*sqs.Message) trace.StartOptions {
		return trace.StartOptions{
			SpanKind: trace.SpanKindClient,
		}
	})
	c := New(nil, WithGetStartOptions(fn))

	fn1 := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	fn2 := runtime.FuncForPC(reflect.ValueOf(c.GetStartOptions).Pointer()).Name()
	assert.Equal(t, fn1, fn2)
}

func TestWithFormatSpanName(t *testing.T) {
	var fn FormatSpanNameFunc = FormatSpanNameFunc(func(*sqs.Message) string {
		return "foo"
	})
	c := New(nil, WithFormatSpanName(fn))

	fn1 := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
	fn2 := runtime.FuncForPC(reflect.ValueOf(c.FormatSpanName).Pointer()).Name()
	assert.Equal(t, fn1, fn2)
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
