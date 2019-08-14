package awsoc // import "go.krak3n.codes/awsoc"

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/aws/aws-sdk-go/service/sqs/sqsiface"
	"go.krak3n.codes/awsoc/propagation"
	"go.krak3n.codes/awsoc/propagation/b3"
	"go.opencensus.io/trace"
)

// A SQSGetStartOptionsFunc returns start options on message by message basis
type SQSGetStartOptionsFunc func(*sqs.Message) trace.StartOptions

// A SQSFormatSpanNameFunc formats a span name from the sqs message
type SQSFormatSpanNameFunc func(*sqs.Message) string

// An SQSOption function customises a clients configuration
type SQSOption func(s *SQS)

// SQSPropagator sets the clients propagator
func SQSPropagator(p propagation.Propagator) SQSOption {
	return SQSOption(func(s *SQS) {
		s.Propagator = p
	})
}

// SQSStartOptions sets the clients StartOptions
func SQSStartOptions(o trace.StartOptions) SQSOption {
	return SQSOption(func(s *SQS) {
		s.StartOptions = o
	})
}

// SQSGetStartOptions sets the SQS clients GetStartOptions func
func SQSGetStartOptions(fn SQSGetStartOptionsFunc) SQSOption {
	return SQSOption(func(s *SQS) {
		s.GetStartOptions = fn
	})
}

// SQSFormatSpanName sets the SQS clients formant name func
func SQSFormatSpanName(fn SQSFormatSpanNameFunc) SQSOption {
	return SQSOption(func(s *SQS) {
		s.FormatSpanName = fn
	})
}

// SQS embeds the AWS SDK SQS client
type SQS struct {
	// Base is an SQS client that satisfies the sqsiface.SQSAPI interface
	Base sqsiface.SQSAPI

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
	GetStartOptions SQSGetStartOptionsFunc

	// FormatSpanName formats the span name based on the given sqs.Message. See
	// DefaultFormatSpanName for the default format
	FormatSpanName SQSFormatSpanNameFunc
}

// NewSQS constructs a new SQS client
func NewSQS(base sqsiface.SQSAPI, opts ...SQSOption) *SQS {
	s := &SQS{
		Base:       base,
		Propagator: b3.New(),
		StartOptions: trace.StartOptions{
			SpanKind: trace.SpanKindServer,
		},
		FormatSpanName: SQSDefaultFormatSpanName,
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
	return send(ctx, s.Base, s.Propagator, in)
}

// sender sends a message to an SQS queue
type sender interface {
	SendMessage(*sqs.SendMessageInput) (*sqs.SendMessageOutput, error)
}

// send sends message to an SQS queue adding span cotnext message attributes for
// propagation according to the given Propagator
func send(ctx aws.Context, sender sender, propagator propagation.Propagator, in *sqs.SendMessageInput) (*sqs.SendMessageOutput, error) {
	if in.MessageAttributes == nil {
		in.MessageAttributes = make(map[string]*sqs.MessageAttributeValue)
	}

	if span := trace.FromContext(ctx); span != nil {
		propagator.SpanContextToMessageAttributes(span.SpanContext(), in.MessageAttributes)
	}

	if in.QueueUrl != nil {
		in.MessageAttributes[TraceQueueURL] = &sqs.MessageAttributeValue{
			StringValue: in.QueueUrl,
			DataType:    aws.String("String"),
		}
	}

	return sender.SendMessage(in)
}

// StartSpanFromMessage a span from an sqs.Message
func (s *SQS) StartSpanFromMessage(ctx context.Context, msg *sqs.Message) (context.Context, *trace.Span) {
	name := s.FormatSpanName(msg)

	sctx, ok := s.Propagator.SpanContextFromMessageAttributes(msg)
	if !ok {
		return trace.StartSpan(ctx, s.FormatSpanName(msg))
	}

	opts := s.StartOptions
	if s.GetStartOptions != nil {
		opts = s.GetStartOptions(msg)
	}

	return trace.StartSpanWithRemoteParent(
		ctx,
		name,
		sctx,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithSampler(opts.Sampler))
}

// SQSDefaultFormatSpanName formats a span name according to the given SQS
// message
// It look at the message attributes for:
// - Queue Url Path
// - Topic Name
// It will also include the Message ID
// Examples:
// - Default: sqs.Message/$MESSAGEID
// - Queue Url: sqs.Message/$QUEUEURLPATH/$MESSAGEID
// - Topic Name: sqs.Message/$TOPICNAME/$MESSAGEID
// - Topic Name and Queue Url Path: sqs.Message/$TOPIC/$QUEUEURLPATH/$MESSAGEID
func SQSDefaultFormatSpanName(msg *sqs.Message) string {
	format := []string{
		"sqs.Message",
		"%s",
	}

	values := []interface{}{
		*msg.MessageId,
	}

	if msg.MessageAttributes != nil {
		var topic string
		if v, ok := msg.MessageAttributes[TraceTopicName]; ok && v.StringValue != nil {
			topic = *v.StringValue
		}

		var queue string
		if v, ok := msg.MessageAttributes[TraceQueueURL]; ok && v.StringValue != nil {
			if u, err := url.Parse(*v.StringValue); err != nil {
				queue = u.Path
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
