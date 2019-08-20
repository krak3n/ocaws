package b3 // import "go.krak3n.codes/ocaws/propagation/b3"

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.opencensus.io/plugin/ochttp/propagation/b3"
	"go.opencensus.io/trace"
)

// Message attribute keys
const (
	TraceIDKey     = "B3-Trace-ID"
	SpanIDKey      = "B3-Span-ID"
	SpanSampledKey = "B3-Span-Sampled"
)

// Attributes is a temporary store for message attributes for translating
// between SQS and SNS message attribute values
type Attributes map[string]string

// Propagator implements the Propagator interface using B3 style formatting to propagate
// Span contexts on SNS / SQS messages
type Propagator struct{}

// New constructs a new B3 based propagator
func New() *Propagator {
	return &Propagator{}
}

// SpanContextToMessageAttributes takes a trace.SpanContext and adds attributes
// to a given sqs / sns message
func (p *Propagator) SpanContextToMessageAttributes(sc trace.SpanContext, t interface{}) bool {
	sampled := "0"
	if sc.IsSampled() {
		sampled = "1"
	}

	attrs := Attributes{
		TraceIDKey:     sc.TraceID.String(),
		SpanIDKey:      sc.SpanID.String(),
		SpanSampledKey: sampled,
	}

	for k, v := range attrs {
		switch T := t.(type) {
		case map[string]*sqs.MessageAttributeValue:
			T[k] = &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v),
			}
		case map[string]*sns.MessageAttributeValue:
			T[k] = &sns.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v),
			}
		default:
			return false
		}
	}

	return true
}

// SpanContextFromMessageAttributes returns a trace.SpanContext based on a sqs
// message
func (p *Propagator) SpanContextFromMessageAttributes(v interface{}) (trace.SpanContext, bool) {
	var (
		tid     trace.TraceID
		sid     trace.SpanID
		sampled trace.TraceOptions
	)

	kv := MessageAttributeValueToAttributes(v)

	if v, ok := kv[TraceIDKey]; ok {
		tid, ok = b3.ParseTraceID(v)
		if !ok {
			return trace.SpanContext{}, false
		}
	} else {
		return trace.SpanContext{}, false
	}

	if v, ok := kv[SpanIDKey]; ok {
		sid, ok = b3.ParseSpanID(v)
		if !ok {
			return trace.SpanContext{}, false
		}
	} else {
		return trace.SpanContext{}, false
	}

	if v, ok := kv[SpanSampledKey]; ok {
		sampled, _ = b3.ParseSampled(v)
	}

	return trace.SpanContext{
		TraceID:      tid,
		SpanID:       sid,
		TraceOptions: sampled,
	}, true
}

// MessageAttributeValueToAttributes converts MessageAttributeValues to key value map
func MessageAttributeValueToAttributes(v interface{}) Attributes {
	attr := Attributes{}

	// TODO: A better way to dry this
	switch t := v.(type) {
	case map[string]*sqs.MessageAttributeValue:
		for k, v := range t {
			if v.StringValue == nil {
				continue
			}
			attr[k] = *v.StringValue
		}
	case map[string]*sns.MessageAttributeValue:
		for k, v := range t {
			if v.StringValue == nil {
				continue
			}
			attr[k] = *v.StringValue
		}
	}

	return attr
}
