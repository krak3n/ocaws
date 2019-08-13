package awsoc // import "github.com/krak3n/awsoc"

import (
	"awsoc/propagation"
	"awsoc/propagation/b3"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/aws/aws-sdk-go/service/sns/snsiface"
	"go.opencensus.io/trace"
)

// An SNSOption function customises a clients configuration
type SNSOption func(*SNS)

// SNSPropagator sets the clients propagator
func SNSPropagator(p propagation.Propagator) SNSOption {
	return SNSOption(func(s *SNS) {
		s.Propagator = p
	})
}

// SNS embeds the AWS SDK SNS client
type SNS struct {
	// Base is an SNS client that satisfies the snsiface.SNSAPI interface
	Base snsiface.SNSAPI

	// Propagator defines how traces will be propagated, if not specified this
	// will be B3
	Propagator propagation.Propagator
}

// New constructs a new SNS client with default configuration values. Use Option
// functions to customise configuration. By default the propagator used is B3.
func New(client snsiface.SNSAPI, opts ...SNSOption) *SNS {
	s := &SNS{
		Base:       client,
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
	return publish(ctx, sns.Base, sns.Propagator, input, opts...)
}

// A publisher publishes messages to SNS
type publisher interface {
	PublishWithContext(ctx aws.Context, input *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error)
}

// publish piblishes messages to SNS
func publish(ctx aws.Context, publisher publisher, propagator propagation.Propagator, in *sns.PublishInput, opts ...request.Option) (*sns.PublishOutput, error) {
	if in.MessageAttributes == nil {
		in.MessageAttributes = make(map[string]*sns.MessageAttributeValue)
	}

	if span := trace.FromContext(ctx); span != nil {
		propagator.SpanContextToMessageAttributes(span.SpanContext(), in.MessageAttributes)
	}

	var arn string
	switch {
	case in.TopicArn != nil:
		arn = *in.TopicArn
	case in.TargetArn != nil:
		arn = *in.TargetArn
	}

	if arn != "" {
		in.MessageAttributes[TraceTopicName] = &sns.MessageAttributeValue{
			StringValue: aws.String(topicNameFromARN(arn)),
			DataType:    aws.String("String"),
		}
	}

	return publisher.PublishWithContext(ctx, in, opts...)
}

// topicNameFromARN grabs the topic name from an ARN, this breaks the ARN at
// : and rteturns the last element of the slice
func topicNameFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) == 0 {
		return ""
	}

	return parts[len(parts)-1]
}
