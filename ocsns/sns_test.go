package ocsns

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/ocawstest"
	"go.krak3n.codes/ocaws/propagation"
	"go.krak3n.codes/ocaws/propagation/propagationtest"
	"go.opencensus.io/trace"
)

type PublishWithContextFunc func(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error)

func (fn PublishWithContextFunc) PublishWithContext(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
	return fn(ctx, input, opts...)
}

func TestMain(m *testing.M) {
	trace.ApplyConfig(trace.Config{
		IDGenerator: ocawstest.NewTestIDGenerator(),
	})

	os.Exit(m.Run())
}

func TestNew(t *testing.T) {
	session, err := session.NewSession(&aws.Config{})
	require.NoError(t, err)

	snsclient := sns.New(session)

	type TestCase struct {
		tName  string
		opts   []Option
		client *SNS
	}
	tt := []TestCase{
		{
			tName: "with propagator",
			opts: []Option{
				WithPropagator(&propagationtest.TestPropator{}),
			},
			client: &SNS{
				SNS:        snsclient,
				Propagator: &propagationtest.TestPropator{},
			},
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			t.Parallel()

			c := New(snsclient, tc.opts...)
			assert.Equal(t, tc.client, c)
		})
	}
}

func Test_publish(t *testing.T) {
	type TestCase struct {
		tName      string
		propagator propagation.Propagator
		publisher  func(*testing.T) publisher
		ctx        context.Context
		in         *sns.PublishInput
		err        error
	}
	tt := []TestCase{
		{
			tName: "no span",
			ctx:   context.Background(),
			publisher: func(t *testing.T) publisher {
				return PublishWithContextFunc(func(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
					assert.Nil(t, input.MessageAttributes)
					return nil, nil
				})
			},
			in: &sns.PublishInput{},
		},
		{
			tName: "with span",
			ctx: func() context.Context {
				ctx, span := trace.StartSpan(context.Background(), t.Name())
				defer span.End()

				return ctx
			}(),
			propagator: &propagationtest.TestPropator{
				SpanContextToMessageAttributesFunc: func(sc trace.SpanContext, v interface{}) bool {
					if T, ok := v.(map[string]*sns.MessageAttributeValue); ok {
						T["TraceID"] = &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(sc.TraceID.String()),
						}
						T["SpanID"] = &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(sc.SpanID.String()),
						}
						T["Sampled"] = &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(fmt.Sprintf("%t", sc.IsSampled())),
						}
					}

					return true
				},
			},
			publisher: func(t *testing.T) publisher {
				return PublishWithContextFunc(func(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
					attr := map[string]*sns.MessageAttributeValue{
						"TraceID": &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(ocawstest.DefaultTraceID.String()),
						},
						"SpanID": &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String(ocawstest.DefaultSpanID.String()),
						},
						"Sampled": &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String("false"),
						},
						ocaws.TraceTopicName: &sns.MessageAttributeValue{
							DataType:    aws.String("String"),
							StringValue: aws.String("Foo"),
						},
					}

					assert.Equal(t, attr, input.MessageAttributes)
					return nil, nil
				})
			},
			in: &sns.PublishInput{
				TopicArn: aws.String("arn:aws:sns:us-east-2:123456789012:Foo"),
			},
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			t.Parallel()

			_, err := publish(tc.ctx, tc.publisher(t), tc.propagator, tc.in)

			assert.Equal(t, tc.err, err)
		})
	}
}

func Test_topicNameFromARN(t *testing.T) {
	type TestCase struct {
		tName string
		arn   string
		topic string
	}
	tt := []TestCase{
		{
			tName: "ok",
			arn:   "arn:aws:sns:us-east-2:123456789012:Foo",
			topic: "Foo",
		},
		{
			tName: "empty",
			arn:   "",
			topic: "",
		},
		{
			tName: "no delimiter",
			arn:   "foo",
			topic: "foo",
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tc.topic, topicNameFromARN(tc.arn))
		})
	}
}
