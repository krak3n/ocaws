package ocsns // import "go.krak3n.codes/ocaws/ocsns"

import (
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sns"
	"go.krak3n.codes/ocaws"
	"go.krak3n.codes/ocaws/propagation"
	"go.krak3n.codes/ocaws/propagation/b3"
	"go.opencensus.io/trace"
)

// An Option function customises a clients configuration
type Option func(*SNS)

// WithPropagator sets the clients propagator
func WithPropagator(p propagation.Propagator) Option {
	return Option(func(s *SNS) {
		s.Propagator = p
	})
}

// SNS embeds the AWS SDK SNS client allowing to be used as a drop in
// replacement for your existing SNS client.
type SNS struct {
	*sns.SNS

	// Propagator defines how traces will be propagated, if not specified this
	// will be B3
	Propagator propagation.Propagator
}

// New constructs a new SNS client with default configuration values. Use
// Option functions to customise configuration. By default the propagator used
// is B3.
func New(client *sns.SNS, opts ...Option) *SNS {
	s := &SNS{
		SNS:        client,
		Propagator: b3.New(),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// PublishWithContext wraps the AWS SDK SNS PublishWithContext method applying
// span context to the input message attributes according the given propagator
func (sns *SNS) PublishWithContext(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
	return publish(ctx, sns.SNS, sns.Propagator, input, opts...)
}

// A publisher publishes messages to SNS
type publisher interface {
	PublishWithContext(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error)
}

// publish piblishes messages to SNS
func publish(ctx aws.Context, publisher publisher, propagator propagation.Propagator, in *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
	if span := trace.FromContext(ctx); span != nil {
		if in.MessageAttributes == nil {
			in.MessageAttributes = make(map[string]*sns.MessageAttributeValue)
		}

		if propagator.SpanContextToMessageAttributes(span.SpanContext(), in.MessageAttributes) {
			if in.TopicArn != nil {
				in.MessageAttributes[ocaws.TraceTopicName] = &sns.MessageAttributeValue{
					StringValue: aws.String(topicNameFromARN(*in.TopicArn)),
					DataType:    aws.String("String"),
				}
			}
		}
	}

	return publisher.PublishWithContext(ctx, in, opts...)
}

// topicNameFromARN grabs the topic name from an ARN, this breaks the ARN at
// : and rteturns the last element of the slice
func topicNameFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	return parts[len(parts)-1]
}
