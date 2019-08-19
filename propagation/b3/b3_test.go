package b3

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/stretchr/testify/assert"
	"go.krak3n.codes/ocaws/ocawstest"
	"go.opencensus.io/trace"
)

func TestSpanContextToMessageAttributes(t *testing.T) {
	type TestCase struct {
		tName    string
		sc       trace.SpanContext
		in       interface{}
		expected interface{}
		ok       bool
	}
	tt := []TestCase{
		{
			tName: "nil",
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(0),
			},
			in: nil,
			ok: false,
		},
		{
			tName: "invalid type",
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(0),
			},
			in:       map[string]string{},
			expected: map[string]string{},
			ok:       false,
		},
		{
			tName: "sns not sampled",
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(0),
			},
			in: map[string]*sns.MessageAttributeValue{},
			expected: map[string]*sns.MessageAttributeValue{
				TraceIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("0"),
				},
			},
			ok: true,
		},
		{
			tName: "sqs not sampled",
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(0),
			},
			in: map[string]*sqs.MessageAttributeValue{},
			expected: map[string]*sqs.MessageAttributeValue{
				TraceIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("0"),
				},
			},
			ok: true,
		},
		{
			tName: "sns sampled",
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(1),
			},
			in: map[string]*sns.MessageAttributeValue{},
			expected: map[string]*sns.MessageAttributeValue{
				TraceIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("1"),
				},
			},
			ok: true,
		},
		{
			tName: "sqs sampled",
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(1),
			},
			in: map[string]*sqs.MessageAttributeValue{},
			expected: map[string]*sqs.MessageAttributeValue{
				TraceIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("1"),
				},
			},
			ok: true,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			t.Parallel()

			p := New()
			ok := p.SpanContextToMessageAttributes(tc.sc, tc.in)

			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.expected, tc.in)
		})
	}
}

func TestSpanContextFromMessageAttributes(t *testing.T) {
	type TestCase struct {
		tName string
		in    interface{}
		sc    trace.SpanContext
		ok    bool
	}
	tt := []TestCase{
		{
			tName: "sns not sampled",
			in: map[string]*sns.MessageAttributeValue{
				TraceIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("0"),
				},
			},
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(0),
			},
			ok: true,
		},
		{
			tName: "sqs not sampled",
			in: map[string]*sqs.MessageAttributeValue{
				TraceIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("0"),
				},
			},
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(0),
			},
			ok: true,
		},
		{
			tName: "sns sampled",
			in: map[string]*sns.MessageAttributeValue{
				TraceIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sns.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("1"),
				},
			},
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(1),
			},
			ok: true,
		},
		{
			tName: "sqs sampled",
			in: map[string]*sqs.MessageAttributeValue{
				TraceIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultTraceID.String()),
				},
				SpanIDKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String(ocawstest.DefaultSpanID.String()),
				},
				SpanSampledKey: &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("1"),
				},
			},
			sc: trace.SpanContext{
				TraceID:      ocawstest.DefaultTraceID,
				SpanID:       ocawstest.DefaultSpanID,
				TraceOptions: trace.TraceOptions(1),
			},
			ok: true,
		},
	}
	for _, tc := range tt {
		tc := tc
		t.Run(tc.tName, func(t *testing.T) {
			t.Parallel()

			p := New()
			sc, ok := p.SpanContextFromMessageAttributes(tc.in)

			assert.Equal(t, tc.ok, ok)
			assert.Equal(t, tc.sc, sc)
		})
	}
}
