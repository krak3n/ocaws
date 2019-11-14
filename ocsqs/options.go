package ocsqs

import (
	"github.com/aws/aws-sdk-go/service/sqs"
	"go.krak3n.codes/ocaws/propagation"
	"go.krak3n.codes/ocaws/propagation/b3"
	"go.opencensus.io/trace"
)

// A GetStartOptionsFunc returns start options on message by message basis
type GetStartOptionsFunc func(*sqs.Message) trace.StartOptions

// A FormatSpanNameFunc formats a span name from the sqs message
type FormatSpanNameFunc func(*sqs.Message) string

type Options struct {
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
}

// DefaultOptions returns sane default options
func DefaultOptions() *Options {
	return &Options{
		Propagator:     b3.New(),
		FormatSpanName: DefaultFormatSpanName,
		StartOptions: trace.StartOptions{
			SpanKind: trace.SpanKindServer,
		},
	}
}

// Option overrides default Options configuration
type Option func(*Options)

// WithPropagator sets the clients propagator
func WithPropagator(p propagation.Propagator) Option {
	return Option(func(o *Options) {
		o.Propagator = p
	})
}

// WithStartOptions sets the clients StartOptions
func WithStartOptions(s trace.StartOptions) Option {
	return Option(func(o *Options) {
		o.StartOptions = s
	})
}

// WithGetStartOptions sets the SQS clients GetStartOptions func
func WithGetStartOptions(fn GetStartOptionsFunc) Option {
	return Option(func(o *Options) {
		o.GetStartOptions = fn
	})
}

// WithFormatSpanName sets the SQS clients formant name func
func WithFormatSpanName(fn FormatSpanNameFunc) Option {
	return Option(func(o *Options) {
		o.FormatSpanName = fn
	})
}
