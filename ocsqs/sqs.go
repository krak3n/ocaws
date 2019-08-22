package ocsqs // import "go.krak3n.codes/ocaws/ocsqs"

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/propagation"
	"go.krak3n.codes/ocaws/propagation/b3"
	"go.opencensus.io/trace"
)

// spanContextKey is the context key for a span context on a context
type spanContextKey = struct{}

// A GetStartOptionsFunc returns start options on message by message basis
type GetStartOptionsFunc func(*sqs.Message) trace.StartOptions

// A FormatSpanNameFunc formats a span name from the sqs message
type FormatSpanNameFunc func(*sqs.Message) string

// An Option function customises a clients configuration
type Option func(s *SQS)

// WithPropagator sets the clients propagator
func WithPropagator(p propagation.Propagator) Option {
	return Option(func(s *SQS) {
		s.Propagator = p
	})
}

// WithStartOptions sets the clients StartOptions
func WithStartOptions(o trace.StartOptions) Option {
	return Option(func(s *SQS) {
		s.StartOptions = o
	})
}

// WithGetStartOptions sets the SQS clients GetStartOptions func
func WithGetStartOptions(fn GetStartOptionsFunc) Option {
	return Option(func(s *SQS) {
		s.GetStartOptions = fn
	})
}

// WithFormatSpanName sets the SQS clients formant name func
func WithFormatSpanName(fn FormatSpanNameFunc) Option {
	return Option(func(s *SQS) {
		s.FormatSpanName = fn
	})
}

// WithRawMessageDelivery sets the SQS client to expect message to be published
// via SNS subscriptions with RawMessageDelivery enabled
func WithRawMessageDelivery() Option {
	return Option(func(s *SQS) {
		s.RawMessageDelivery = true
	})
}

// SQSAPI embeds the sqsiface.SQSAPI interface and extends it to include methods
// for sending messages with span context and starting spans from messages.
type SQSAPI interface {
	sqsiface.SQSAPI

	SendMessageContext(ctx aws.Context, in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
	StartSpanFromMessage(ctx context.Context, msg *sqs.Message) (context.Context, *trace.Span)
	ContextWithSpanFromMessage(ctx context.Context, msg *sqs.Message) context.Context
}

// SQS provides methods for sending messages with trace attributes and starting
// spans from messages. It embeds the SQS client allowing this to be used as
// a drop in replacement.
type SQS struct {
	*sqs.SQS

	// Propagator defines how traces will be propagated, if not specified this
	// will be B3
	Propagator propagation.Propagator

	// StartOptions are applied to the span started by this Handler around each
	// message.
	// StartOptions.SpanKind will always be set to trace.SpanKindServer
	// for spans started by this transport.
	StartOptions trace.StartOptions

	// GetStartOptions allows to set start options per message. If set,
	// StartOptions is going to be ignored.
	GetStartOptions GetStartOptionsFunc

	// FormatSpanName formats the span name based on the given sqs.Message. See
	// DefaultFormatSpanName for the default format
	FormatSpanName FormatSpanNameFunc

	// If you have setup your SQS SNS subscription to use RawMessageDelivery you
	// should enable this using the WithRawMessageDelivery. The raw message
	// attributes will be passed to the Propagator directly rather than
	// unmarshalling the message body to build the message attrubutes
	RawMessageDelivery bool
}

// New constructs a new SQS client with default configuration values. Use
// Option functions to customise configuration. By default the propagator used
// is B3.
func New(client *sqs.SQS, opts ...Option) *SQS {
	s := &SQS{
		SQS:        client,
		Propagator: b3.New(),
		StartOptions: trace.StartOptions{
			SpanKind: trace.SpanKindServer,
		},
		FormatSpanName: DefaultFormatSpanName,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// SendMessageContext sends a message to SQS propagating span context on the
// message attributes.
// Note: This method does not currently exist on the SQS client unlike the SNS
// client, this is a best guess at the API looking at existing AWS SDK patterns
// where context is included, under the hood this will call sqs.SendMessage
func (s *SQS) SendMessageContext(ctx aws.Context, in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	return send(ctx, s.SQS, s.Propagator, in)
}

// sender sends a message to an SQS queue
type sender interface {
	SendMessageRequest(*sqs.SendMessageInput) (*request.Request, *sqs.SendMessageOutput)
}

// send sends message to an SQS queue adding span cotnext message attributes for
// propagation according to the given Propagator
func send(ctx aws.Context, sender sender, propagator propagation.Propagator, in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	if span := trace.FromContext(ctx); span != nil {
		if in.MessageAttributes == nil {
			in.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
		}

		if ok := propagator.SpanContextToMessageAttributes(span.SpanContext(), in.MessageAttributes); ok {
			if in.QueueUrl != nil {
				in.MessageAttributes[ocaws.TraceQueueURL] = &sqs.MessageAttributeValue{
					StringValue: in.QueueUrl,
					DataType:    aws.String("String"),
				}
			}
		}
	}

	req, out := sender.SendMessageRequest(in)
	req.SetContext(ctx)

	if err := req.Send(); err != nil {
		return nil, err
	}

	return out, nil
}

// StartSpanFromMessage a span from an sqs.Message
func (s *SQS) StartSpanFromMessage(ctx context.Context, msg *sqs.Message) (context.Context, *trace.Span) {
	sctx, ok := s.Propagator.SpanContextFromMessageAttributes(s.getMessageAttributes(msg))
	if !ok {
		return trace.StartSpan(ctx, s.FormatSpanName(msg))
	}

	opts := s.StartOptions
	if s.GetStartOptions != nil {
		opts = s.GetStartOptions(msg)
	}

	return trace.StartSpanWithRemoteParent(
		ctx,
		s.FormatSpanName(msg),
		sctx,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithSampler(opts.Sampler))
}

// ContextWithSpanFromMessage will add a span context from a message onto the given
// context retuning a new context, this allows for defered starting of spans
func (s *SQS) ContextWithSpanFromMessage(ctx context.Context, msg *sqs.Message) context.Context {
	sctx, ok := s.Propagator.SpanContextFromMessageAttributes(s.getMessageAttributes(msg))
	if !ok {
		return ctx
	}

	return context.WithValue(ctx, spanContextKey{}, sctx)
}

// getMessageAttributes returns message attributes from an SQS message, if
// RawMessageDelivery is enabled the message attributes are returned else the
// message body is unmarshaled and message attributes are built from the body
func (s *SQS) getMessageAttributes(msg *sqs.Message) map[string]*sqs.MessageAttributeValue {
	if s.RawMessageDelivery {
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

// SpanFromContext will return a span context from context
func SpanFromContext(ctx context.Context) (trace.SpanContext, bool) {
	v, ok := ctx.Value(spanContextKey{}).(trace.SpanContext)
	return v, ok
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
