package ocsqs // import "go.krak3n.codes/ocaws/ocsqs"

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/sqs"
)

// SQS provides methods for sending messages with trace attributes and starting
// spans from messages. It embeds the SQS client allowing this to be used as
// a drop in replacement.
type SQS struct {
	*sqs.SQS

	options []Option
}

// New constructs a new SQS client with default configuration values. Use
// Option functions to customize configuration. By default the propagator used
// is B3.
func New(client *sqs.SQS, opts ...Option) *SQS {
	return &SQS{
		client,
		opts,
	}
}

// SendMessageWithContext shadows the sqs clients SendMessageWithContext adding trace span data to
// the send message input
func (s *SQS) SendMessageWithContext(ctx aws.Context, input *sqs.SendMessageInput, opts ...request.Option) (*sqs.SendMessageOutput, error) {
	return s.SQS.SendMessageWithContext(ctx, SendMessageInputWithSpan(ctx, input, s.options...), opts...)
}
