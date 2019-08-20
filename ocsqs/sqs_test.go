package ocsqs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/ocawstest"
	"go.krak3n.codes/ocaws/propagation"
	"go.krak3n.codes/ocaws/propagation/propagationtest"
	"go.opencensus.io/trace"
)

type SendMessageRequestFunc func(*sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput)

func (fn SendMessageRequestFunc) SendMessageRequest(in *sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
	return fn(in)
}

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

func Test_send(t *testing.T) {
	type TestCase struct {
		tName      string
		sender     func(*testing.T, string) sender
		propagator propagation.Propagator
		handler    func(*testing.T) http.Handler
		ctx        context.Context
		in         *sqs.SendMessageInput
		attributes map[string]*sqs.MessageAttributeValue
		err        error
	}
	tt := []TestCase{
		{
			tName: "req error",
			ctx:   context.Background(),
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Helper()

					w.WriteHeader(http.StatusInternalServerError)
				})
			},
			sender: func(t *testing.T, url string) sender {
				return SendMessageRequestFunc(func(in *sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
					t.Helper()

					r, err := http.NewRequest(http.MethodGet, url, nil)
					require.NoError(t, err)

					return &request.Request{
						HTTPRequest: r,
						Error:       errors.New("boom"),
					}, nil
				})
			},
			in:  &sqs.SendMessageInput{},
			err: errors.New("boom"),
		},
		{
			tName: "no span",
			ctx:   context.Background(),
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Helper()

					w.WriteHeader(http.StatusOK)
				})
			},
			sender: func(t *testing.T, url string) sender {
				return SendMessageRequestFunc(func(in *sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
					t.Helper()

					r, err := http.NewRequest(http.MethodGet, url, nil)
					require.NoError(t, err)

					return &request.Request{
						HTTPRequest: r,
					}, nil
				})
			},
			in: &sqs.SendMessageInput{
				QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/123456789101112/Foo"),
			},
		},
		{
			tName: "no queue url",
			ctx: func() context.Context {
				ctx, span := trace.StartSpan(context.Background(), t.Name())
				defer span.End()

				return ctx
			}(),
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Helper()

					w.WriteHeader(http.StatusOK)
				})
			},
			sender: func(t *testing.T, url string) sender {
				return SendMessageRequestFunc(func(in *sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
					t.Helper()

					r, err := http.NewRequest(http.MethodGet, url, nil)
					require.NoError(t, err)

					return &request.Request{
						HTTPRequest: r,
					}, nil
				})
			},
			propagator: &propagationtest.TestPropator{
				SpanContextToMessageAttributesFunc: func(sc trace.SpanContext, v interface{}) bool {
					if T, ok := v.(map[string]*sqs.MessageAttributeValue); ok {
						T["TraceID"] = &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(sc.TraceID.String()),
						}
						T["SpanID"] = &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(sc.SpanID.String()),
						}
						T["Sampled"] = &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(fmt.Sprintf("%t", sc.IsSampled())),
						}
					}
					return true
				},
			},
			in: &sqs.SendMessageInput{},
			attributes: map[string]*sqs.MessageAttributeValue{
				"TraceID": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				"SpanID": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				"Sampled": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("false"),
				},
			},
		},
		{
			tName: "ok",
			ctx: func() context.Context {
				ctx, span := trace.StartSpan(context.Background(), t.Name())
				defer span.End()

				return ctx
			}(),
			handler: func(t *testing.T) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					t.Helper()

					w.WriteHeader(http.StatusOK)
				})
			},
			sender: func(t *testing.T, url string) sender {
				return SendMessageRequestFunc(func(in *sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput) {
					t.Helper()

					r, err := http.NewRequest(http.MethodGet, url, nil)
					require.NoError(t, err)

					return &request.Request{
						HTTPRequest: r,
					}, nil
				})
			},
			propagator: &propagationtest.TestPropator{
				SpanContextToMessageAttributesFunc: func(sc trace.SpanContext, v interface{}) bool {
					if T, ok := v.(map[string]*sqs.MessageAttributeValue); ok {
						T["TraceID"] = &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(sc.TraceID.String()),
						}
						T["SpanID"] = &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(sc.SpanID.String()),
						}
						T["Sampled"] = &sqs.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(fmt.Sprintf("%t", sc.IsSampled())),
						}
					}
					return true
				},
			},
			in: &sqs.SendMessageInput{
				QueueUrl: aws.String("https://sqs.eu-west-1.amazonaws.com/123456789101112/Foo"),
			},
			attributes: map[string]*sqs.MessageAttributeValue{
				"TraceID": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				"SpanID": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				"Sampled": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("false"),
				},
				ocaws.TraceQueueURL: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("https://sqs.eu-west-1.amazonaws.com/123456789101112/Foo"),
				},
			},
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			t.Parallel()

			srv := httptest.NewServer(tc.handler(t))
			defer srv.Close()

			_, err := send(tc.ctx, tc.sender(t, srv.URL), tc.propagator, tc.in)
			assert.Equal(t, tc.err, err)
			assert.Equal(t, tc.attributes, tc.in.MessageAttributes)
		})
	}
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
