package propagation // import "github.com/krak3n/awsoc/propagation"

import "go.opencensus.io/trace"

// A Propagator propagates span context to and from message attributes.
// Due to the way the AWS SDK types message attributes (each package has their
// own type of message attribute) we cannot create an interface around this, so
// we need to take empty interfaces and type switch in the propagator
// implementations
type Propagator interface {
	SpanContextToMessageAttributes(sc trace.SpanContext, v interface{}) bool
	SpanContextFromMessageAttributes(v interface{}) (trace.SpanContext, bool)
}
