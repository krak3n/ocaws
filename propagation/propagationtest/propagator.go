package propagationtest

import "go.opencensus.io/trace"

// TestPropator is a test propagator
type TestPropator struct {
	SpanContextToMessageAttributesFunc   func(sc trace.SpanContext, v interface{}) bool
	SpanContextFromMessageAttributesFunc func(v interface{}) (trace.SpanContext, bool)
}

// SpanContextToMessageAttributes adds messages attributes from a span
func (t *TestPropator) SpanContextToMessageAttributes(sc trace.SpanContext, v interface{}) bool {
	if t.SpanContextToMessageAttributesFunc != nil {
		return t.SpanContextToMessageAttributesFunc(sc, v)
	}

	return false
}

// SpanContextFromMessageAttributes returns a span context form message attributes
func (t *TestPropator) SpanContextFromMessageAttributes(v interface{}) (trace.SpanContext, bool) {
	if t.SpanContextFromMessageAttributesFunc != nil {
		return t.SpanContextFromMessageAttributesFunc(v)
	}

	return trace.SpanContext{}, false
}
