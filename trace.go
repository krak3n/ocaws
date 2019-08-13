package awsoc

import (
	"context"

	"go.opencensus.io/trace"
)

// SpanContextFromContext returns a trace.SpanContext from a context
func SpanContextFromContext(ctx context.Context) (trace.SpanContext, bool) {
	return trace.SpanContext{}, false
}

// ContextWithSpanContext stores a trace.SpanContext on a context returning the
// new context
func ContextWithSpanContext(ctx context.Context, sc trace.SpanContext) context.Context {
	return ctx
}
