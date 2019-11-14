package ocsqs

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.krak3n.codes/ocaws"
	"go.opencensus.io/trace"
)

// spanContextKey is the context key for a span context on a context
type spanContextKey = struct{}

// StartSpan starts a span from an SQS Message
func StartSpan(ctx context.Context, msg *sqs.Message, opts ...Option) (context.Context, *trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	o := DefaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	name := o.FormatSpanName(msg)
	attrs := GetMessageAttributes(msg)

	sctx, ok := o.Propagator.SpanContextFromMessageAttributes(attrs)
	if !ok {
		return trace.StartSpan(ctx, name)
	}

	sopts := o.StartOptions
	if o.GetStartOptions != nil {
		sopts = o.GetStartOptions(msg)
	}

	return trace.StartSpanWithRemoteParent(
		ctx,
		name,
		sctx,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithSampler(sopts.Sampler))
}

// WithContext will create a new span context and place it on the given context from a message. This
// is useful if you wish to defer the starting of a span
func WithContext(ctx context.Context, msg *sqs.Message, opts ...Option) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	o := DefaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	attrs := GetMessageAttributes(msg)

	sctx, ok := o.Propagator.SpanContextFromMessageAttributes(attrs)
	if !ok {
		return ctx
	}

	return context.WithValue(ctx, spanContextKey{}, sctx)
}

// SendMessageInputWithSpan adds span data to message input to propagate spans being send through
// SQS directly.
func SendMessageInputWithSpan(ctx context.Context, in *sqs.SendMessageInput, opts ...Option) *sqs.SendMessageInput {
	if ctx == nil {
		return in
	}

	o := DefaultOptions()
	for _, opt := range opts {
		opt(o)
	}

	in2 := new(sqs.SendMessageInput)
	*in = *in2

	if span := trace.FromContext(ctx); span != nil {
		if in2.MessageAttributes == nil {
			in2.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
		}

		if ok := o.Propagator.SpanContextToMessageAttributes(span.SpanContext(), in2.MessageAttributes); ok {
			if in2.QueueUrl != nil {
				in2.MessageAttributes[ocaws.TraceQueueURL] = &sqs.MessageAttributeValue{
					StringValue: in2.QueueUrl,
					DataType:    aws.String("String"),
				}
			}
		}
	}

	return in2
}

// GetMessageAttributes returns message attributes from an SQS message
func GetMessageAttributes(msg *sqs.Message) map[string]*sqs.MessageAttributeValue {
	if msg.MessageAttributes != nil {
		return msg.MessageAttributes
	}

	attr := map[string]*sqs.MessageAttributeValue{}
	if msg.Body == nil {
		return nil
	}

	dst := map[string]json.RawMessage{}
	if err := json.Unmarshal([]byte(*msg.Body), &dst); err != nil {
		return nil
	}

	if v, ok := dst["MessageAttributes"]; ok {
		dst := map[string]map[string]string{}
		if err := json.Unmarshal(v, &dst); err != nil {
			return nil
		}

		for k, v := range dst {
			attr[k] = &sqs.MessageAttributeValue{
				DataType:    aws.String("String"),
				StringValue: aws.String(v["Value"]),
			}
		}
	}

	return attr
}

// DefaultFormatSpanName formats a span name according to the given SQS
// message.
func DefaultFormatSpanName(msg *sqs.Message) string {
	format := []string{
		"sqs.Message",
		"%s",
	}

	mid := "unknwonMessageId"
	if msg.MessageId != nil {
		mid = *msg.MessageId
	}

	values := []interface{}{
		mid,
	}

	if msg.MessageAttributes != nil {
		var topic string
		if v, ok := msg.MessageAttributes[ocaws.TraceTopicName]; ok && v.StringValue != nil {
			topic = *v.StringValue
		}

		var queue string
		if v, ok := msg.MessageAttributes[ocaws.TraceQueueURL]; ok && v.StringValue != nil {
			if u, err := url.Parse(*v.StringValue); err == nil {
				queue = strings.TrimLeft(u.Path, "/")
			}
		}

		switch {
		case (topic != "" && queue != ""):
			format = append(format, "%s", "%s")
			values = append([]interface{}{topic, queue}, values...)
		case topic != "":
			format = append(format, "%s")
			values = append([]interface{}{topic}, values...)
		case queue != "":
			format = append(format, "%s")
			values = append([]interface{}{queue}, values...)
		}
	}

	return fmt.Sprintf(strings.Join(format, "/"), values...)
}

// SpanFromContext will return a span context from context
func SpanFromContext(ctx context.Context) (trace.SpanContext, bool) {
	v, ok := ctx.Value(spanContextKey{}).(trace.SpanContext)
	return v, ok
}
